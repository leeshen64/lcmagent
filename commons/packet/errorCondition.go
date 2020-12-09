package packet

type ErrorCondition string

const (
	E_BAD_REQUEST                 ErrorCondition = "E_BAD_REQUEST"
	E_CONFLICT                    ErrorCondition = "E_CONFLICT"
	E_FEATURE_NOT_IMPLEMENTED     ErrorCondition = "E_FEATURE_NOT_IMPLEMENTED"
	E_FORBIDDEN                   ErrorCondition = "E_FORBIDDEN"
	E_INTERNAL_SERVER_ERROR       ErrorCondition = "E_INTERNAL_SERVER_ERROR"
	E_ITEM_NOT_FOUND              ErrorCondition = "E_ITEM_NOT_FOUND"
	E_NOT_ACCEPTABLE              ErrorCondition = "E_NOT_ACCEPTABLE"
	E_NOT_ALLOWED                 ErrorCondition = "E_NOT_ALLOWED"
	E_NOT_AUTHORIZED              ErrorCondition = "E_NOT_AUTHORIZED"
	E_REGISTRATION_REQUIRED       ErrorCondition = "E_REGISTRATION_REQUIRED"
	E_REMOTE_SERVER_NOT_AVAILABLE ErrorCondition = "E_REMOTE_SERVER_NOT_AVAILABLE"
	E_REMOTE_SERVER_TIMEOUT       ErrorCondition = "E_REMOTE_SERVER_TIMEOUT"
	E_SERVICE_UNAVAILABLE         ErrorCondition = "E_SERVICE_UNAVAILABLE"
	E_UNEXPECTED_CONDITION        ErrorCondition = "E_UNEXPECTED_CONDITION"
)

// String : convert element to string
func (e ErrorCondition) String() string {
	switch e {
	case E_BAD_REQUEST:
		return "E_BAD_REQUEST"
	case E_CONFLICT:
		return "E_CONFLICT"
	case E_FEATURE_NOT_IMPLEMENTED:
		return "E_FEATURE_NOT_IMPLEMENTED"
	case E_FORBIDDEN:
		return "E_FORBIDDEN"
	case E_INTERNAL_SERVER_ERROR:
		return "E_INTERNAL_SERVER_ERROR"
	case E_ITEM_NOT_FOUND:
		return "E_ITEM_NOT_FOUND"
	case E_NOT_ACCEPTABLE:
		return "E_NOT_ACCEPTABLE"
	case E_NOT_ALLOWED:
		return "E_NOT_ALLOWED"
	case E_NOT_AUTHORIZED:
		return "E_NOT_AUTHORIZED"
	case E_REGISTRATION_REQUIRED:
		return "E_REGISTRATION_REQUIRED"
	case E_REMOTE_SERVER_NOT_AVAILABLE:
		return "E_REMOTE_SERVER_NOT_AVAILABLE"
	case E_REMOTE_SERVER_TIMEOUT:
		return "E_REMOTE_SERVER_TIMEOUT"
	case E_SERVICE_UNAVAILABLE:
		return "E_SERVICE_UNAVAILABLE"
	case E_UNEXPECTED_CONDITION:
		return "E_UNEXPECTED_CONDITION"
	default:
		return ""
	}
}

func ParseCondition(conditionString string) ErrorCondition {
	switch conditionString {
	case "E_BAD_REQUEST":
		return E_BAD_REQUEST
	case "E_CONFLICT":
		return E_CONFLICT
	case "E_FEATURE_NOT_IMPLEMENTED":
		return E_FEATURE_NOT_IMPLEMENTED
	case "E_FORBIDDEN":
		return E_FORBIDDEN
	case "E_INTERNAL_SERVER_ERROR":
		return E_INTERNAL_SERVER_ERROR
	case "E_ITEM_NOT_FOUND":
		return E_ITEM_NOT_FOUND
	case "E_NOT_ACCEPTABLE":
		return E_NOT_ACCEPTABLE
	case "E_NOT_ALLOWED":
		return E_NOT_ALLOWED
	case "E_NOT_AUTHORIZED":
		return E_NOT_AUTHORIZED
	case "E_REGISTRATION_REQUIRED":
		return E_REGISTRATION_REQUIRED
	case "E_REMOTE_SERVER_NOT_AVAILABLE":
		return E_REMOTE_SERVER_NOT_AVAILABLE
	case "E_REMOTE_SERVER_TIMEOUT":
		return E_REMOTE_SERVER_TIMEOUT
	case "E_SERVICE_UNAVAILABLE":
		return E_SERVICE_UNAVAILABLE
	case "E_UNEXPECTED_CONDITION":
		return E_UNEXPECTED_CONDITION
	default:
		return ""
	}
}
