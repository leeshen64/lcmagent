package ws

import (
	"sercomm.com/demeter/commons/packet"
)

// ServerSessionHandler ...
type ServerSessionHandler interface {
	SessionCreated(session *ServerSession)
	SessionDestroyed(session *ServerSession)
	SessionMessageReceived(session *ServerSession, datagram *packet.Datagram)
	SessionHeartbeat(session *ServerSession)
}
