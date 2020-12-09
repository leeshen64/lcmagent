package ws

import (
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/zap"
	"sercomm.com/demeter/commons/logger"
	"sercomm.com/demeter/commons/packet"
)

// ServerSession ...
type ServerSession struct {
	Session

	id           string
	connection   *websocket.Conn      // websocket connection
	connectionID int64                // ID of the websocket connection assigned by server
	handler      ServerSessionHandler // interface is a pointer. https://stackoverflow.com/questions/44370277/type-is-pointer-to-interface-not-interface-confusion
	state        SessionState         // session state
	locker       *sync.Mutex          // prevent "write message" parallelly
	asyncCallMap cmap.ConcurrentMap   // pair< datagram id, asyncCall object >
	propertyMap  cmap.ConcurrentMap   // pair <property key, property value>, session's local properties
}

// NewServerSession ...
func NewServerSession(sessionID string, connection *websocket.Conn, sessionHandler ServerSessionHandler) *ServerSession {
	if nil == connection {
		panic("'connection' CANNOT BE nil")
	}

	if nil == sessionHandler {
		panic("'sessionHandler' CANNOT BE nil")
	}

	connectionID := atomic.AddInt64(&sessionIDCounter, 1)
	serverSession := &ServerSession{
		id:           sessionID,
		connection:   connection,
		connectionID: connectionID,
		handler:      sessionHandler,
		state:        StateConnected} // session has arrived and connected

	connection.SetCloseHandler(serverSession.closeHandler)

	return serverSession
}

// GetID ...
// (public) Implementation of Session interface
func (serverSession *ServerSession) GetID() string {
	return serverSession.id
}

// GetConnection ...
// (public) Implementation of Session interface
func (serverSession *ServerSession) GetConnection() *websocket.Conn {
	return serverSession.connection
}

// GetState ...
// (public) Implementation of Session interface
func (serverSession *ServerSession) GetState() SessionState {
	return serverSession.state
}

// Close ...
// (public) Implementation of Session interface
func (serverSession *ServerSession) Close(statusCode int, reason string) error {
	if serverSession.state == StateClosed {
		return nil
	}
	// update state to closed
	serverSession.state = StateClosed

	if nil != serverSession.handler {
		defer serverSession.handler.SessionDestroyed(serverSession)
	}

	if nil != serverSession.connection {
		err := serverSession.connection.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(statusCode, reason))
		if err != nil {
			return err
		}

		// close the connection
		err = serverSession.connection.Close()
		serverSession.connection = nil
		return err
	}

	return nil
}

// Serve ...
// (public)
func (serverSession *ServerSession) Serve() error {
	if serverSession.handler != nil {
		serverSession.handler.SessionCreated(serverSession)
	}

	for {
		messageType, message, err := serverSession.GetConnection().ReadMessage()
		if err == nil {
			serverSession.messageHandler(messageType, message)
		} else {
			if serverSession.state == StateConnected {
				serverSession.Close(websocket.CloseAbnormalClosure, "ERROR WHILE READING MESSAGES")
				return err
			}

			return nil
		}
	}
}

// GetProperty ...
func (serverSession *ServerSession) GetProperty(key string) interface{} {
	value, _ := serverSession.propertyMap.Get(key)
	return value
}

// SetProperty ...
func (serverSession *ServerSession) SetProperty(key string, value interface{}) {
	serverSession.propertyMap.Set(key, value)
}

// handle close
func (serverSession *ServerSession) closeHandler(code int, reason string) error {
	// update state
	serverSession.state = StateClosed

	// trigger handler.SessionDestroyed
	if serverSession.handler != nil {
		defer serverSession.handler.SessionDestroyed(serverSession)
	}

	// destroy connection
	serverSession.connection.Close()
	serverSession.connection = nil

	return nil
}

// handle message
func (serverSession *ServerSession) messageHandler(messageType int, message []byte) {
	if messageType != websocket.TextMessage {
		serverSession.Close(websocket.CloseInvalidFramePayloadData, "ONLY TEXT PAYLOAD WILL BE ACCEPTED")
	}

	// convert byte message to string
	messageString := string(message)
	logger.New().Info("websocket: RECV", zap.String("MESSAGE", messageString))

	if messageString == "" {
		// ignore blank message
		return
	}

	// parse datagram
	datagram, err := packet.From(messageString)
	if nil != err {
		logger.New().Error("websocket: PARSE", zap.String("MESSAGE", messageString), zap.Error(err))
		return
	}

	// remove async call object of the datagram if exists
	defer serverSession.asyncCallMap.Remove(datagram.ID)

	typeValue := packet.ParseType(datagram.Type)

	switch typeValue {
	case packet.T_REQUEST:
		// trigger handler.SessionReceiveRequest
		if nil != serverSession.handler {
			serverSession.handler.SessionMessageReceived(serverSession, &datagram)
		}
		return
	case packet.T_RESULT:
		// load callback object from async call map
		object, ok := serverSession.asyncCallMap.Get(datagram.ID)
		if true == ok {
			asyncCallObject := object.(asyncCall)
			asyncCallObject.Timeout.Stop()
			if nil != asyncCallObject.OnResult {
				asyncCallObject.OnResult(serverSession, datagram.ID, datagram.Arguments...)
			}
		}
		return
	case packet.T_ERROR:
		// load callback function from async call map
		object, ok := serverSession.asyncCallMap.Get(datagram.ID)
		if true == ok {
			asyncCallObject := object.(asyncCall)
			asyncCallObject.Timeout.Stop()

			errorCondition := packet.ParseCondition(datagram.Arguments[0].(string))
			errorMessage := datagram.Arguments[1].(string)
			if nil != asyncCallObject.OnError {
				asyncCallObject.OnError(serverSession, datagram.ID, errorCondition, errorMessage)
			}
		}
		return
	default:
		errorDatagram := packet.Datagram{
			ID:       datagram.ID,
			Type:     packet.T_ERROR.String(),
			Function: datagram.Function,
		}

		errorDatagram.Push(packet.E_BAD_REQUEST)
		errorDatagram.Push("UNKNOWN DATAGRAM TYPE -> " + datagram.Type)

		jsonString, _ := packet.String(&errorDatagram)
		serverSession.connection.WriteMessage(websocket.TextMessage, []byte(jsonString))
		return
	}
}
