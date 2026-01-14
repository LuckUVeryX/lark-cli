package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSON outputs data as JSON to stdout
func JSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// Error outputs an error in JSON format
func Error(code, message string) {
	JSON(map[string]interface{}{
		"error":   true,
		"code":    code,
		"message": message,
	})
}

// ErrorFromErr outputs an error from a Go error
func ErrorFromErr(code string, err error) {
	Error(code, err.Error())
}

// Success outputs a success message
func Success(message string) {
	JSON(map[string]interface{}{
		"success": true,
		"message": message,
	})
}

// Fatal outputs an error and exits with code 1
func Fatal(code string, err error) {
	Error(code, err.Error())
	os.Exit(1)
}

// Fatalf outputs a formatted error and exits
func Fatalf(code, format string, args ...interface{}) {
	Error(code, fmt.Sprintf(format, args...))
	os.Exit(1)
}
