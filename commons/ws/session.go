package ws

import (
	"github.com/gorilla/websocket"
	"sercomm.com/demeter/commons/packet"
)

// SessionState ...
type SessionState int

// StateClosed : session closed
// StateConnecting : session is connecting
// StateConnected : session was connected
const (
	StateClosed     SessionState = 0
	StateConnecting SessionState = 1
	StateConnected  SessionState = 2
)

var sessionIDCounter int64

// Session ...
type Session interface {
	GetID() string
	GetConnection() *websocket.Conn
	GetState() SessionState
	Close(statusCode int, reason string) error
	Deliver(
		datagram *packet.Datagram,
		timeoutInterval int,
		onResult func(session Session, packetID string, arguments ...interface{}),
		onError func(session Session, packetID string, condition packet.ErrorCondition, errorMessage string),
		onTimeout func(session Session, packetID string, timeoutInterval int)) error
}
