package ws

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http/httputil"
	"net/url"
	"runtime/debug"
	"strconv"
	"sync"

	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/zap"
	"sercomm.com/demeter/commons/logger"
	"sercomm.com/demeter/commons/packet"
	"sercomm.com/demeter/commons/util"
)

// ClientSession ...
type ClientSession struct {
	Session

	id           string
	connection   *websocket.Conn
	handler      ClientSessionHandler // interface is a pointer. https://stackoverflow.com/questions/44370277/type-is-pointer-to-interface-not-interface-confusion
	state        SessionState         // session state
	locker       *sync.Mutex          // prevent "write message" parallelly
	ping         Ping                 // timer for "ping"
	asyncCallMap cmap.ConcurrentMap   // pair< datagram id, asyncCall object >
}

// NewClientSession ...
func NewClientSession(sessionHandler ClientSessionHandler) *ClientSession {
	if nil == sessionHandler {
		panic("'sessionHandler' CANNOT BE NIL")
	}

	id := atomic.AddInt64(&sessionIDCounter, 1)
	idString := strconv.FormatInt(id, 10)
	clientSession := &ClientSession{
		id:           idString,
		connection:   nil,
		handler:      sessionHandler,
		state:        StateClosed,
		asyncCallMap: cmap.New()}

	clientSession.locker = &sync.Mutex{}

	return clientSession
}

// GetID is the implementation of Session.GetID()
func (clientSession *ClientSession) GetID() string {
	return clientSession.id
}

// GetConnection is the implementation of Session.GetConnection()
func (clientSession *ClientSession) GetConnection() *websocket.Conn {
	return clientSession.connection
}

// GetState is the implementation of Session.GetState()
func (clientSession *ClientSession) GetState() SessionState {
	return clientSession.state
}

// Open ...
func (clientSession *ClientSession) Open(host string, port int, path string, pingPeriod int, tlsConfig *tls.Config) error {
	var err error

	if StateClosed != clientSession.state {
		err = errors.New("ALREADY OPENED")
		return err
	}

	clientSession.state = StateConnecting

	var scheme string
	if nil != tlsConfig {
		scheme = "wss"
		websocket.DefaultDialer.TLSClientConfig = tlsConfig
	} else {
		scheme = "ws"
	}

	address := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   path,
	}
	// connect to server
	connection, resp, err := websocket.DefaultDialer.Dial(address.String(), nil)
	if nil != err {
		if err == websocket.ErrBadHandshake {
			request, _ := httputil.DumpRequestOut(resp.Request, true)
			response, _ := httputil.DumpResponse(resp, true)
			logger.New().Debug("WEBSOCKET HANDSHAKING FAILED", zap.String("HTTP REQUEST", string(request)), zap.String("HTTP RESPONSE", string(response)))
		}

		clientSession.connection = nil
		clientSession.state = StateClosed
		return err
	}

	clientSession.connection = connection
	clientSession.connection.SetCloseHandler(clientSession.closeHandler)
	clientSession.state = StateConnected

	// initialize ping mechanism
	clientSession.ping.retry = false
	clientSession.ping.timer = time.NewTimer(time.Duration(pingPeriod) * time.Second)
	defer clientSession.ping.timer.Stop()
	defer clientSession.connection.Close()
	defer debug.FreeOSMemory()

	var datagram *packet.Datagram = nil

	// ping routine
	go func(rTimer *time.Timer, rSession *ClientSession) {
		for {
			<-rTimer.C
			debug.FreeOSMemory()

			// generate ping datagram
			if nil == datagram {
				datagram = &packet.Datagram{
					ID:       util.RandomUUIDString(),
					Type:     packet.T_REQUEST.String(),
					Function: packet.F_PING.String(),
				}
			} else {
				datagram.ID = util.RandomUUIDString()
			}

			rSession.Deliver(datagram, 15,
				func(session Session, datagramID string, arguments ...interface{}) {
					logger.New().Debug("PING SUCCESS")

					rSession.ping.retry = false
					rTimer.Reset(time.Duration(pingPeriod) * time.Second)
				},
				func(session Session, datagramID string, condition packet.ErrorCondition, errorMessage string) {
					logger.New().Debug("PING ERROR", zap.String("CONDITION", condition.String()), zap.String("MESSAGE", errorMessage))

					if rSession.ping.retry == false {
						rSession.ping.retry = true
						rTimer.Reset(time.Duration(pingPeriod) * time.Second)
					} else {
						rSession.Close(websocket.CloseNormalClosure, "PING ERROR")
					}
				},
				func(session Session, datagramID string, timeoutInterval int) {
					logger.New().Debug("PING TIMEOUT")

					if rSession.ping.retry == false {
						rSession.ping.retry = true
						rTimer.Reset(time.Duration(pingPeriod) * time.Second)
					} else {
						rSession.Close(websocket.CloseNormalClosure, "PING TIMEOUT")
					}
				})
		}
	}(clientSession.ping.timer, clientSession)

	// trigger handler.SessionCreated
	clientSession.handler.SessionCreated(clientSession)

	// read message
	for {
		messageType, message, err := clientSession.connection.ReadMessage()
		if err != nil {
			if nil != clientSession.connection {
				if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					// update state
					clientSession.state = StateClosed
					// trigger handler.SessionDestroyed
					clientSession.handler.SessionDestroyed(clientSession)
				}
			}
			break
		}

		clientSession.messageHandler(messageType, message)
	}

	return err
}

// Close ...
func (clientSession *ClientSession) Close(statusCode int, reason string) error {
	if clientSession.state == StateClosed {
		return nil
	}
	// update state to closed
	clientSession.state = StateClosed

	if nil != clientSession.handler {
		defer clientSession.handler.SessionDestroyed(clientSession)
	}

	connection := clientSession.connection
	if nil != connection {
		// send close message
		clientSession.connection.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(statusCode, reason),
			time.Now().Add(time.Duration(3000)*time.Millisecond))

		defer func() {
			clientSession.connection = nil
		}()

		return clientSession.connection.Close()
	}

	return nil
}

// Deliver ...
// (public) Deliver message to server
func (clientSession *ClientSession) Deliver(
	datagram *packet.Datagram,
	timeoutInterval int,
	onResult func(session Session, packetID string, arguments ...interface{}),
	onError func(session Session, packetID string, condition packet.ErrorCondition, errorMessage string),
	onTimeout func(session Session, packetID string, timeoutInterval int)) error {

	clientSession.locker.Lock()
	defer clientSession.locker.Unlock()

	if StateConnected != clientSession.GetState() {
		return errors.New("SESSION IS NOT READY")
	}

	var err error = nil

	if datagram.Type == packet.T_REQUEST.String() {
		timer := time.AfterFunc(time.Duration(timeoutInterval)*time.Second, func() {
			clientSession.asyncCallMap.Remove(datagram.ID)

			if nil != onTimeout {
				onTimeout(clientSession, datagram.ID, timeoutInterval)
			}
		})

		// allocate async call object
		asyncCall := asyncCall{
			Timeout:  timer,
			OnResult: onResult,
			OnError:  onError}

		clientSession.asyncCallMap.Set(datagram.ID, asyncCall)
	}

	jsonString, err := packet.String(datagram)
	if nil != err {
		return err
	}

	logger.New().Info("websocket: SEND", zap.String("MESSAGE", jsonString))
	return clientSession.connection.WriteMessage(websocket.TextMessage, []byte(jsonString))
}

// handle close
func (clientSession *ClientSession) closeHandler(statusCode int, reason string) error {
	// update state
	clientSession.state = StateClosed

	// trigger handler.SessionDestroyed
	if nil != clientSession.handler {
		defer clientSession.handler.SessionDestroyed(clientSession)
	}

	// stop the ping timer if necessary
	if nil != clientSession.ping.timer {
		clientSession.ping.timer.Stop()
	}

	return nil
}

// handle message
func (clientSession *ClientSession) messageHandler(messageType int, message []byte) {
	if messageType != websocket.TextMessage {
		clientSession.Close(websocket.CloseInvalidFramePayloadData, "ONLY TEXT PAYLOAD WILL BE ACCEPTED")
	}

	messageString := string(message)
	logger.New().Info("websocket: RECV", zap.String("MESSAGE", messageString))

	// parse datagram
	datagram, err := packet.From(messageString)
	if nil != err {
		logger.New().Error("websocket: PARSE", zap.String("MESSAGE", messageString), zap.Error(err))
		return
	}

	// remove async call object of the datagram if exists
	defer func() {
		object, ok := clientSession.asyncCallMap.Get(datagram.ID)
		if true == ok {
			asyncCallObject := object.(asyncCall)
			asyncCallObject.Timeout.Stop()

			clientSession.asyncCallMap.Remove(datagram.ID)
		}
	}()

	typeValue := packet.ParseType(datagram.Type)

	switch typeValue {
	case packet.T_REQUEST:
		// trigger handler.SessionReceiveRequest
		if nil != clientSession.handler {
			clientSession.handler.SessionMessageReceived(clientSession, &datagram)
		}
		return
	case packet.T_RESULT:
		// load callback function from async call map
		object, ok := clientSession.asyncCallMap.Get(datagram.ID)
		if true == ok {
			asyncCallObject := object.(asyncCall)
			if nil != asyncCallObject.OnResult {
				asyncCallObject.OnResult(clientSession, datagram.ID, datagram.Arguments...)
			}
		}
		return
	case packet.T_ERROR:
		// load callback function from async call map
		object, ok := clientSession.asyncCallMap.Get(datagram.ID)
		if true == ok {
			asyncCallObject := object.(asyncCall)

			errorCondition := packet.ParseCondition(datagram.Arguments[0].(string))
			errorMessage := datagram.Arguments[1].(string)
			if nil != asyncCallObject.OnError {
				asyncCallObject.OnError(clientSession, datagram.ID, errorCondition, errorMessage)
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
		clientSession.connection.WriteMessage(websocket.TextMessage, []byte(jsonString))
		return
	}
}
