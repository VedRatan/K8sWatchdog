package gemini

import "strings"

// Helper function to extract YAML from the Gemini response
func extractYAMLFromResponse(response string) string {
	// Define the start and end markers for the YAML block
	startMarker := "```yaml"
	endMarker := "```"

	// Find the start and end positions of the YAML block
	startIndex := strings.Index(response, startMarker)
	if startIndex == -1 {
		return ""
	}
	startIndex += len(startMarker)

	endIndex := strings.Index(response[startIndex:], endMarker)
	if endIndex == -1 {
		return ""
	}
	endIndex += startIndex

	// Extract the YAML content
	yamlContent := response[startIndex:endIndex]

	// Trim any leading or trailing whitespace
	yamlContent = strings.TrimSpace(yamlContent)

	return yamlContent
}
