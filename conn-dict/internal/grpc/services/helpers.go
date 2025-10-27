package services

// Helper functions shared across service implementations

// getStringOrEmpty safely extracts a string value from a map or returns empty string
func getStringOrEmpty(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
