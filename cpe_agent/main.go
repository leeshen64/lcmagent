package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"os"
	"os/signal"
	"syscall"
	"time"

	//"net/http"
	//_ "net/http/pprof"

	"github.com/gorilla/websocket"
	//_ "github.com/mkevac/debugcharts"
	daemon "github.com/sevlyar/go-daemon"

	"go.uber.org/zap"
	"sercomm.com/demeter/commons/configger"
	"sercomm.com/demeter/commons/logger"
	utility "sercomm.com/demeter/commons/util"
	"sercomm.com/demeter/commons/ws"
	"sercomm.com/demeter/cpe_agent/rpc/caller"
	"sercomm.com/demeter/cpe_agent/rpc/model"
)

// BUILD_TIME timestamp when building cpe_agent
var BUILD_TIME string = "N/A"

// VERSION version information
var VERSION string = fmt.Sprintf("V1.0.16 %s", BUILD_TIME)

// CONF_PATH1 1st default configuration file path
const CONF_PATH1 string = string(os.PathSeparator) + "conf" + string(os.PathSeparator) + "cpe_agent.yaml"

// CONF_PATH2 2nd default configuration file path
const CONF_PATH2 string = string(os.PathSeparator) + "etc" + string(os.PathSeparator) + "cpe_agent.yaml"

var context *daemon.Context = nil
var session *ws.ClientSession = nil
var certPool *x509.CertPool = nil
var isReady bool = false
var hardwareInfo model.SystemHardware

func main() {
	var showHelp bool
	var showVersion bool
	var daemonMode bool
	var confPath string
	var standaloneMode bool
	flag.BoolVar(&showHelp, "h", false, "help")
	flag.BoolVar(&showVersion, "v", false, "version")
	flag.BoolVar(&daemonMode, "d", false, "daemon mode")
	flag.StringVar(&confPath, "c", "", "configuration file path (must be in YAML format)")
	flag.BoolVar(&standaloneMode, "s", false, "standalone mode with fake hardware information")
	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	// working directory
	directory, err := utility.GetCurrentDirectory()
	if nil != err {
		fmt.Println("UNABLE TO DETECT WORKING DIRECTORY: " + err.Error())
		os.Exit(1)
	}

	// load configuration
	if confPath == "" {
		fmt.Println("NO CONFIGURATION FILE SPECIFIED, LOADING DEFAULT CONFIGURATION...")
		confPath = directory + CONF_PATH1
		err = configger.Load(confPath)
		if nil != err {
			confPath = CONF_PATH2
			err = configger.Load(confPath)
		}
	} else {
		err = configger.Load(confPath)
	}

	if nil != err {
		fmt.Println("UNABLE TO LOAD CONFIGURATION FILE: " + err.Error())
		os.Exit(1)
	}

	// enable daemon mode
	if daemonMode {
		context, err = daemonize(standaloneMode)
		if nil != err {
			os.Exit(1)
		}
		defer context.Release()
	}

	/*
		go func() {
			// terminal: $ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap
			// web:
			// 1、http://localhost:8081/ui
			// 2、http://localhost:6060/debug/charts
			// 3、http://localhost:6060/debug/pprof
			http.ListenAndServe("0.0.0.0:6060", nil)
		}()
	*/

	logFoler := configger.GetValue("log", "folder")
	logRotateCount := configger.GetValue("log", "rotateCount")

	logger.SetEnvParam(logFoler.String(directory), logRotateCount.Int(2))
	logger.New().Info("START PROC: " + VERSION)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	terminated := false

	hostValue := configger.GetValue("entry", "host")
	portValue := configger.GetValue("entry", "port")
	pathValue := configger.GetValue("entry", "path")
	enableSSLValue := configger.GetValue("entry", "enableSSL")
	pingValue := configger.GetValue("entry", "pingPeriod")

	host := hostValue.String("localhost")
	port := portValue.Int(443)
	path := pathValue.String("/iface/v1/cpe")
	enableSSL := enableSSLValue.Bool(true)
	ping := pingValue.Int(30)

	var tlsConfig *tls.Config
	if enableSSL == true {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		tlsConfig = nil
	}

	go func() {
		defer func() {
			if nil != session {
				session.Close(websocket.CloseNormalClosure, "")
			}
		}()

		for {
			if true == terminated {
				break
			}

			if isReady == false {
				err = queryHardwareInformation(standaloneMode)
				if nil != err {
					logger.New().Error("UNABLE TO LOAD HARDWARE INFORMATION: " + err.Error())
				} else {
					logger.New().Info("System.Hardware",
						zap.String("SerialNumber", hardwareInfo.SerialNumber),
						zap.String("MACAddress", hardwareInfo.MAC),
						zap.String("ModelName", hardwareInfo.Model),
						zap.String("FirmwareVersion", hardwareInfo.SoftwareVersion))
				}
			}

			if isReady == true && (session == nil || session.GetState() != ws.StateConnected) {
				if nil == session {
					session = ws.NewClientSession(&CpeSessionHandler{})
				}

				logger.New().Info("websocket: CONNECTING TO HOST...", zap.String("host", host), zap.Int("port", port), zap.String("path", path), zap.Bool("enableSSL", enableSSL), zap.Int("pingPeriod", ping))

				err := session.Open(
					host,
					port,
					path,
					ping,
					tlsConfig)

				if nil != err {
					logger.New().Error(err.Error())
				}
			}

			time.Sleep(time.Duration(3) * time.Second)
		}
	}()

	for {
		select {
		case sig := <-interrupt:
			terminated = true
			if nil != session && session.GetState() != ws.StateClosed {
				session.Close(websocket.CloseNormalClosure, "PROCESS INTERRUPTED")
			}

			logger.New().Warn("PROGRAM EXIT: " + sig.String())
			os.Exit(0)
		}
	}
}

func daemonize(standalone bool) (*daemon.Context, error) {
	appName := utility.GetAppName()
	directory, _ := utility.GetCurrentDirectory()

	/*
		var args []string
		if standalone {
			args = []string{appName + "d", "-s"}
		} else {
			args = []string{appName + "d"}
		}
	*/

	context := &daemon.Context{
		PidFileName: "/var/run/" + appName + "d.pid",
		PidFilePerm: 0644,
		LogFileName: "",
		LogFilePerm: 0640,
		WorkDir:     directory,
		Umask:       027,
		Args:        os.Args,
	}

	child, err := context.Reborn()
	if err != nil {
		logger.New().Error("UNABLE TO RUN DAEMONIZE: " + err.Error())
		return context, err
	}

	if child != nil {
		return context, errors.New("FAILED TO CREATE CHILD PROCESS")
	}

	return context, nil
}

func queryHardwareInformation(standalone bool) error {
	var err error

	response := model.SystemHardwareResponse{}

	if false == standalone {
		jsonString, err := caller.Call("Get", "System.Hardware", "")

		if nil == err {
			err = json.Unmarshal([]byte(jsonString), &response)
			if nil == err {
				isReady = true
				hardwareInfo = response.Body
			}
		}
	} else {
		isReady = true
		err = nil

		hardwareInfo = model.SystemHardware{
			ProductClass:    "Sample Device",
			FriendlyName:    "SERCOMM Sample Device",
			Manufacturer:    "SERCOMM",
			Model:           "HG5244B",
			Variant:         "SERCOMM",
			CasingColour:    "Black",
			MAC:             "AABBCCDDEEFF",
			SerialNumber:    "AAAAA00001",
			Carrier:         "SERCOMM",
			SoftwareVersion: "FAKE.1.2.3",
		}
	}

	return err
}
