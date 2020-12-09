package main

import (
	"encoding/json"

	"go.uber.org/zap"
	"sercomm.com/demeter/commons/logger"
	"sercomm.com/demeter/commons/packet"
	"sercomm.com/demeter/commons/util"
	"sercomm.com/demeter/commons/ws"
	"sercomm.com/demeter/cpe_agent/rpc/caller"
)

const ackTimeout = 15

// CpeSessionHandler ...
type CpeSessionHandler struct {
	ws.ClientSessionHandler
}

// SessionCreated ...
func (handler *CpeSessionHandler) SessionCreated(session *ws.ClientSession) {
	logger.New().Info("websocket: SESSION CREATED")

	// raise an identification runtime
	go func() {
		datagram := packet.Datagram{
			ID:       util.RandomUUIDString(),
			Type:     packet.T_REQUEST.String(),
			Function: packet.F_IDENTIFY.String(),
		}

		datagram.Push(hardwareInfo.SerialNumber)
		datagram.Push(hardwareInfo.MAC)
		datagram.Push(hardwareInfo.Model)
		datagram.Push(hardwareInfo.SoftwareVersion)

		// deliver identification packet
		session.Deliver(&datagram, ackTimeout,
			func(session ws.Session, packetID string, arguments ...interface{}) {
				logger.New().Info("IDENTIFICATION SUCCESS")
			},
			func(session ws.Session, packetID string, condition packet.ErrorCondition, errorMessage string) {
				logger.New().Info("IDENTIFICATION FAILURE", zap.String("REASON", errorMessage))
			},
			func(session ws.Session, packetID string, timeoutInterval int) {
				logger.New().Info("IDENTIFICATION TIMEOUT", zap.Int("INTERVAL", timeoutInterval))
			})
	}()
}

// SessionMessageReceived ...
func (handler *CpeSessionHandler) SessionMessageReceived(session *ws.ClientSession, datagram *packet.Datagram) {
	logger.New().Info("websocket: SESSION RECV REQUEST", zap.String("datagramID", datagram.ID))

	function := packet.ParseFunction(datagram.Function)

	// if received F_UNKNOWN
	if function == packet.F_UNKNOWN {
		errorDatagram := &packet.Datagram{
			ID:       datagram.ID,
			Type:     packet.T_ERROR.String(),
			Function: datagram.Function,
		}

		errorDatagram.Push(packet.E_FEATURE_NOT_IMPLEMENTED)
		errorDatagram.Push("UNKNOWN FUNCTION")

		session.Deliver(errorDatagram, 0, nil, nil, nil)
		return
	}

	switch function {
	case packet.F_REBOOT:
		return
	case packet.F_UPGRADE:
		return
	case packet.F_UBUS:
		methodString := util.GetAsString(datagram.Arguments, 0, "")
		pathString := util.GetAsString(datagram.Arguments, 1, "")
		requestString := util.GetAsString(datagram.Arguments, 2, "")
		processUbusCommand(
			session,
			datagram.ID,
			methodString,
			pathString,
			requestString)
		return
	default:
		return
	}

}

// SessionDestroyed ...
func (handler *CpeSessionHandler) SessionDestroyed(session *ws.ClientSession) {
	logger.New().Info("websocket: SESSION DESTROYED")
}

// process ubus command from server
func processUbusCommand(session ws.Session, id string, methodString string, pathString string, requestString string) {
	jsonString, err := caller.Call(methodString, pathString, requestString)

	var datagram *packet.Datagram

	if nil != err || jsonString == "" {
		datagram = &packet.Datagram{
			ID:       id,
			Type:     packet.T_ERROR.String(),
			Function: packet.F_UBUS.String(),
		}

		datagram.Push(packet.E_REMOTE_SERVER_NOT_AVAILABLE)
		if nil != err {
			datagram.Push(err.Error())
		} else {
			datagram.Push("BLANK RESPONSE FROM UBUS")
		}

		// deliver the error to server
		err = session.Deliver(datagram, 0, nil, nil, nil)
		if nil != err {
			logger.New().Warn("CANNOT DELIVER RESULT", zap.String("id", id), zap.Error(err))
		}

		datagram = nil

		// exit the function directly
		return
	}

	// unmarshal the result to server
	var dataModel interface{}
	err = json.Unmarshal([]byte(jsonString), &dataModel)
	if nil != err {
		datagram = &packet.Datagram{
			ID:       id,
			Type:     packet.T_ERROR.String(),
			Function: packet.F_UBUS.String(),
		}

		// error: the result JSON is invalid
		datagram.Push(packet.E_REMOTE_SERVER_NOT_AVAILABLE)
		datagram.Push(err.Error())
	} else {
		datagram = &packet.Datagram{
			ID:       id,
			Type:     packet.T_RESULT.String(),
			Function: packet.F_UBUS.String(),
		}

		datagram.Push(dataModel)
	}

	// deliver the result to server
	err = session.Deliver(datagram, 0, nil, nil, nil)
	if nil != err {
		logger.New().Warn("CANNOT DELIVER RESULT", zap.String("id", id), zap.Error(err))
	}
}
