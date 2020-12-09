package ws

import (
	"time"
)

const pingInterval int = 30

// Ping ...
type Ping struct {
	timer *time.Timer
	retry bool
}
