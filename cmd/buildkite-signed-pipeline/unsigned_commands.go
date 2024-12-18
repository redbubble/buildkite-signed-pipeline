package main

import (
	"strings"
	"path/filepath"
	"runtime"
	"fmt"
	"os"
)

const (
	posixSpecialChars = "!\"#$&'()*,;<=>?[]\\^`{}|~"
	batchSpecialChars = "^&;,=%"
)

func getToolNames() []string {
	thisTool := filepath.Base(os.Args[0])
	toolNames := []string{thisTool}

	// handle both thisTool and thisTool.exe on windows
	if runtime.GOOS == `windows` {
		toolNames = append(toolNames, strings.TrimSuffix(thisTool, ".exe"))
	}

	return toolNames
}

func isUploadCommand(command string) bool {
	for _, toolName := range getToolNames() {
		// buildkite-signed-pipeline upload
		rawUploadCommand := fmt.Sprintf("%s upload", toolName)
		if strings.HasPrefix(command, rawUploadCommand) {
			return true
		}
	}

	// Something we use at Redbubble 😬
	if command == "./script/upload_pipeline" {
		return true
	}
	if command == "script/upload_pipeline" {
		return true
	}

	// vanilla upload command
	if strings.HasPrefix(command, "buildkite-agent pipeline upload") {
		return true
	}

	return false
}

func hasSpecialShellChars(str string) bool {
	if runtime.GOOS == `windows` {
		return strings.ContainsAny(str, batchSpecialChars);
	}
	return strings.ContainsAny(str, posixSpecialChars);
}

func IsUnsignedCommandOk(command string) (bool, error) {
	if !isUploadCommand(command) {
		return false, nil
	}
	// ensure no special shell variables are used, this means `buildkite-agent pipeline upload `rm -rf /`` would be disallowed
	return !hasSpecialShellChars(command), nil
}
