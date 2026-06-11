package git

import (
	"fmt"
	"os"
	"path/filepath"
	"project-service/internal/framework"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
)

func CloneRepository(repoURL string) (string, string, error) {

	if err := ValidateGitHubRepository(repoURL); err != nil {
		return "", "", err
	}

	repoName := getRepositoryName(repoURL)

	baseDir := "repositories"

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create repositories directory: %w", err)
	}

	targetDir := filepath.Join(
		baseDir,
		fmt.Sprintf("%s-%d", repoName, time.Now().Unix()),
	)

	_, err := gogit.PlainClone(
		targetDir,
		false,
		&gogit.CloneOptions{
			URL:      repoURL,
			Progress: os.Stdout,
			Depth:    1,
		},
	)

	if err != nil {
		return "", "", fmt.Errorf("failed to clone repository: %w", err)
	}

	frameworkName, err := framework.DetectFramework(targetDir)
	if err != nil {
		return "", "", fmt.Errorf("framework detection failed: %w", err)
	}

	return targetDir, frameworkName, nil
}
func getRepositoryName(repoURL string) string {

	repoURL = strings.TrimSuffix(repoURL, "/")

	parts := strings.Split(repoURL, "/")

	if len(parts) == 0 {
		return "repository"
	}

	name := parts[len(parts)-1]

	name = strings.TrimSuffix(name, ".git")

	return name
}
