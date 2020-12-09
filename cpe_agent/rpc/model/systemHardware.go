package model

type Header struct {
	Code int
	Name string
}

type SystemHardware struct {
	ProductClass    string
	FriendlyName    string
	Manufacturer    string
	Model           string
	Variant         string
	CasingColour    string
	MAC             string
	SerialNumber    string
	Carrier         string
	SoftwareVersion string
}

type SystemHardwareResponse struct {
	Response Header
	Body     SystemHardware
}
