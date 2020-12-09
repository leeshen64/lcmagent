package packet

type Function string

const (
	F_UNKNOWN  Function = ""
	F_PING     Function = "F_PING"
	F_IDENTIFY Function = "F_IDENTIFY"
	F_REBOOT   Function = "F_REBOOT"
	F_UPGRADE  Function = "F_UPGRADE"
	F_UBUS     Function = "F_UBUS"
)

// String : convert element to string
func (e Function) String() string {
	switch e {
	case F_PING:
		return "F_PING"
	case F_IDENTIFY:
		return "F_IDENTIFY"
	case F_REBOOT:
		return "F_REBOOT"
	case F_UPGRADE:
		return "F_UPGRADE"
	case F_UBUS:
		return "F_UBUS"
	default:
		return ""
	}
}

func ParseFunction(functionString string) Function {
	switch functionString {
	case "F_PING":
		return F_PING
	case "F_IDENTIFY":
		return F_IDENTIFY
	case "F_REBOOT":
		return F_REBOOT
	case "F_UPGRADE":
		return F_UPGRADE
	case "F_UBUS":
		return F_UBUS
	default:
		return F_UNKNOWN
	}
}
