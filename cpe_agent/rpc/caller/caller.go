package caller

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"runtime/debug"
	"time"

	"go.uber.org/zap"
	"sercomm.com/demeter/commons/logger"
)

// Call ... Execute ubus CLI
//    method   - method of target ubus path depends on definitions of each ubus command. E.g. "Get","Install","Delete" etc.
//    path     - ubus command path. E.g. "Services.Management.LCM.ExecutionEnvironments"
//    payload  - payload JSON string. Please refer to Sercomm_LCM_UBUS_API.xlsx or any up-to-date document
func Call(method string, path string, payloadString string) (string, error) {
	//logger.New().Info("UBUS CALL: ", zap.String("PATH", path), zap.String("METHOD", method), zap.String("PAYLOAD", payloadString))

	var responseString string
	var err error

	var cmd *exec.Cmd
	if "" != payloadString {
		cmd = exec.Command("ubus", "call", path, method, payloadString)
	} else {
		cmd = exec.Command("ubus", "call", path, method)
	}

	stdout, err := cmd.StdoutPipe()
	defer stdout.Close()
	if err != nil {
		logger.New().Info("UBUS ERROR: ", zap.String("MESSAGE", err.Error()))
		return responseString, err
	}

	stderr, err := cmd.StderrPipe()
	defer stderr.Close()
	if err != nil {
		logger.New().Info("UBUS ERROR: ", zap.String("MESSAGE", err.Error()))
		return responseString, err
	}

	// execute the command
	if err := cmd.Start(); err != nil {
		logger.New().Info("UBUS ERROR: ", zap.String("MESSAGE", err.Error()))
		return responseString, err
	}

	// read stderr
	buffer, err := ioutil.ReadAll(stderr)

	if nil == err {
		errMsg := string(buffer)
		if errMsg != "" {
			logger.New().Info("UBUS RESPONSE: ", zap.String("ERROR", errMsg))
			err = errors.New(errMsg)
		} else {
			// read stdout
			buffer, err = ioutil.ReadAll(stdout)
			if nil == err {
				responseString = string(buffer)
				logger.New().Info("UBUS RESPONSE: ", zap.String("RESPONSE", responseString))
			}
		}
	}

	// force killing the process if it cannot exit normally
	timer := time.AfterFunc(time.Duration(3)*time.Second, func() {
		if err := cmd.Process.Kill(); err != nil {
			logger.New().Info("UBUS FAILED TO KILL PROCESS: ", zap.String("MESSAGE", err.Error()))
		}

		logger.New().Info("UBUS PROCESS KILLED AS TIMEOUT REACHED")
	})

	// wait for the process to finish or kill it after 3 seconds (whichever happens first):
	done := make(chan error, 1)
	defer close(done)

	go func() {
		done <- cmd.Wait()
		defer cmd.Process.Release()
		defer timer.Stop()
		defer debug.FreeOSMemory()
	}()

	select {
	case err := <-done:
		if err != nil {
			logger.New().Info("UBUS PROCESS FINISHED WITH ERROR: ", zap.String("MESSAGE", err.Error()))
		}
	}

	return responseString, err
}
