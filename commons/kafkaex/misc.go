package kafkaex

import (
	"encoding/binary"
)

// int64ToBytes is unsafe for singed integer value
// since UNIX epoch time has no sign, it is for this package internally
func int64ToBytes(value int64) []byte {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}

// int64ToBytes is unsafe for singed integer value
// since UNIX epoch time has no sign, it is for this package internally
func bytesToInt64(buf []byte) int64 {
	return int64(binary.LittleEndian.Uint64(buf))
}
