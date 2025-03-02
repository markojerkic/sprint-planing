package components

// CardProps defines properties for the Card component
type CardProps struct {
	HasAccent   bool   // Add the accent border
	ClassName   string // Additional custom classes
	HeaderClass string // Additional header classes
	BodyClass   string // Additional body classes
	Title       string // Card title
	Subtitle    string // Card subtitle
}

// AlertType defines the type of alert
type AlertType string

const (
	Success AlertType = "success"
	Warning AlertType = "warning"
	Error   AlertType = "error"
	Info    AlertType = "info"
)

// Helper functions
func Ternary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func TernaryStr(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
