package util

import (
	"encoding/json"
	"log/slog"

	"github.com/labstack/echo/v4"
)

func AddTostHeader(c echo.Context, text string) error {
	// Get the current HX-Trigger header, if any
	var triggers map[string]any
	currentTrigger := c.Response().Header().Get("HX-Trigger")

	if currentTrigger != "" {
		// Parse existing triggers
		if err := json.Unmarshal([]byte(currentTrigger), &triggers); err != nil {
			slog.Error("Error parsing existing HX-Trigger header", slog.Any("error", err))
			// Initialize a new map if we couldn't parse the existing one
			triggers = make(map[string]any)
		}
	} else {
		// Initialize a new map if there was no existing header
		triggers = make(map[string]any)
	}

	// Add toast event to the triggers
	triggers["toast"] = map[string]any{
		"message": text,
	}

	// Marshal back to JSON
	jsonData, err := json.Marshal(triggers)
	if err != nil {
		slog.Error("Error encoding toast message", slog.Any("error", err))
		return err
	}

	slog.Debug("Setting updated toast header", slog.String("toast", string(jsonData)))
	c.Response().Header().Set("HX-Trigger", string(jsonData))
	return nil
}
