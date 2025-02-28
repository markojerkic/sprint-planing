package components

import "github.com/a-h/templ"

// ButtonProps defines properties for the Button component
type ButtonProps struct {
	Type      string                // "submit", "button", "reset"
	ClassName string                // Additional custom classes
	IsBlock   bool                  // Full width button
	IsLarge   bool                  // Larger size
	ID        string                // Button ID attribute
	Disabled  bool                  // Disabled state
	FormID    string                // Form attribute
	AriaLabel string                // Accessibility label
	OnClick   templ.ComponentScript // JavaScript onClick handler
}

// InputProps defines properties for the FormInput component
type InputProps struct {
	Type        string // "text", "email", "password", etc.
	ID          string // Input ID
	Name        string // Input name
	Label       string // Input label
	Placeholder string // Placeholder text
	Value       string // Input value
	Required    bool   // Required attribute
	Disabled    bool   // Disabled state
	MaxLength   int    // Maximum character length
	Pattern     string // Validation pattern
	HelpText    string // Help text below input
	ClassName   string // Additional custom classes
	LabelClass  string // Additional label classes
	ErrorMsg    string // Error message
	HasError    bool   // Error state
}

// CardProps defines properties for the Card component
type CardProps struct {
	HasAccent   bool   // Add the accent border
	ClassName   string // Additional custom classes
	HeaderClass string // Additional header classes
	BodyClass   string // Additional body classes
	Title       string // Card title
	Subtitle    string // Card subtitle
}

// FormProps defines properties for the Form component
type FormProps struct {
	ID         string                // Form ID
	Action     string                // Form action URL
	Method     string                // Form method
	ClassName  string                // Additional custom classes
	OnSubmit   templ.ComponentScript // JavaScript onSubmit handler
	HasEnctype bool                  // Whether to add enctype for file uploads
	NoValidate bool                  // Whether to disable browser validation
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
