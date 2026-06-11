package framework

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	FrameworkGo      = "go"
	FrameworkNodeJS  = "nodejs"
	FrameworkPython  = "python"
	FrameworkJava    = "java"
	FrameworkUnknown = "unknown"
)

// DetectFramework detects the project framework based on common files.
func DetectFramework(repoPath string) (string, error) {

	info, err := os.Stat(repoPath)
	if err != nil {
		return FrameworkUnknown, fmt.Errorf("repository path does not exist: %w", err)
	}

	if !info.IsDir() {
		return FrameworkUnknown, fmt.Errorf("repository path is not a directory")
	}

	switch {
	case fileExists(filepath.Join(repoPath, "go.mod")):
		return FrameworkGo, nil

	case fileExists(filepath.Join(repoPath, "package.json")):
		return FrameworkNodeJS, nil

	case fileExists(filepath.Join(repoPath, "requirements.txt")):
		return FrameworkPython, nil

	case fileExists(filepath.Join(repoPath, "pom.xml")):
		return FrameworkJava, nil

	case fileExists(filepath.Join(repoPath, "build.gradle")):
		return FrameworkJava, nil

	case fileExists(filepath.Join(repoPath, "build.gradle.kts")):
		return FrameworkJava, nil

	default:
		return FrameworkUnknown, nil
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
