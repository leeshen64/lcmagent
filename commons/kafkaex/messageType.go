package kafkaex

type MessageType string

const (
	MessageTypeInvalid  MessageType = ""
	MessageTypeRequest  MessageType = "Request"
	MessageTypeResponse MessageType = "Response"
)

func (e MessageType) String() string {
	return string(e)
}

func ParseMessageTypeString(value string) MessageType {
	switch value {
	case MessageTypeRequest.String():
		return MessageTypeRequest
	case MessageTypeResponse.String():
		return MessageTypeResponse
	}

	return MessageTypeInvalid
}
