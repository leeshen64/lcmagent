package packet

type Type string

const (
	T_REQUEST Type = "T_REQUEST"
	T_RESULT  Type = "T_RESULT"
	T_ERROR   Type = "T_ERROR"
)

// String : convert element to string
func (e Type) String() string {
	switch e {
	case T_REQUEST:
		return "T_REQUEST"
	case T_RESULT:
		return "T_RESULT"
	case T_ERROR:
		return "T_ERROR"
	default:
		return ""
	}
}

func ParseType(typeString string) Type {
	switch typeString {
	case "T_REQUEST":
		return T_REQUEST
	case "T_RESULT":
		return T_RESULT
	case "T_ERROR":
		return T_ERROR
	default:
		return ""
	}
}
