package types

// Request body for POST /send
type SendSMSRequest struct {
	Port     string `json:"port"`       // e.g. "/dev/ttyUSB0" hoặc "COM3"
	BaudRate int    `json:"baud_rate"`  // default 115200
	To       string `json:"to"`         // số điện thoại
	Message  string `json:"message"`    // nội dung SMS
	Timeout  int    `json:"timeout"`    // timeout in seconds (default 30)
}
