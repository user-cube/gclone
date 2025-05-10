package ui

// OperationInfo displays information about an operation
func OperationInfo(operation string, profile string, details map[string]string) {
	Section(operation + " with profile: " + Highlight(profile))

	for key, value := range details {
		PrintKeyValue(key, value)
	}

	Normal("\n")
}

// OperationSuccess displays a success message for an operation
func OperationSuccess(message string, details ...string) {
	Success("%s\n", message)
	for _, detail := range details {
		Normal("  %s\n", detail)
	}
}

// OperationError displays an error message for an operation
func OperationError(operation string, err error) {
	Error("Error %s: %v\n", operation, err)
}
