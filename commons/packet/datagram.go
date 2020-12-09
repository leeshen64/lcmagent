package packet

import (
	"encoding/json"
)

// Datagram : structure of Datagram
type Datagram struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Function  string        `json:"function"`
	Arguments []interface{} `json:"arguments"`
}

// Push : push argument
func (datagram *Datagram) Push(argument interface{}) {
	datagram.Arguments = append(datagram.Arguments, argument)
}

// From : convert from JSON string
func From(jsonString string) (Datagram, error) {
	var packet Datagram
	err := json.Unmarshal([]byte(jsonString), &packet)
	return packet, err
}

// String : convert to JSON string
func String(packet *Datagram) (string, error) {
	var jsonString string
	var jsonData []byte
	var err error
	jsonData, err = json.Marshal(packet)
	if nil == err {
		jsonString = string(jsonData)
	}

	return jsonString, err
}
