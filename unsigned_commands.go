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

func isUploadCommand(command string) bool {
	// buildkite-signed-pipeline upload
	rawUploadCommand := fmt.Sprintf("%s upload", filepath.Base(os.Args[0]))
	if strings.HasPrefix(command, rawUploadCommand) {
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
