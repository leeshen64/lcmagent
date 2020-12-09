package ws

import (
	"sercomm.com/demeter/commons/packet"
	"time"
)

type asyncCall struct {
	Timeout  *time.Timer
	OnResult func(session Session, datagramID string, arguments ...interface{})
	OnError  func(session Session, datagramID string, condition packet.ErrorCondition, errorMessage string)
}
