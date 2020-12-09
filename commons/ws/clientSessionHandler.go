package ws

import (
	"sercomm.com/demeter/commons/packet"
)

// ClientSessionHandler ...
type ClientSessionHandler interface {
	SessionCreated(session *ClientSession)
	SessionDestroyed(session *ClientSession)
	SessionMessageReceived(session *ClientSession, datagram *packet.Datagram)
}
