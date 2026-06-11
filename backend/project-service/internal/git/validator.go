package git

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GitHubRepository struct {
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

func ValidateGitHubRepository(repoURL string) error {

	u, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("invalid url")
	}

	if u.Scheme != "https" {
		return fmt.Errorf("github url must use https")
	}

	if u.Host != "github.com" {
		return fmt.Errorf("only github.com repositories are supported")
	}

	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		return fmt.Errorf("invalid github repository format")
	}

	owner := parts[0]
	repo := parts[1]

	apiURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s",
		owner,
		repo,
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request")
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "mini-kubernetes-platform")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to reach github")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("repository does not exist")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github returned status %d", resp.StatusCode)
	}

	var repository GitHubRepository

	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return fmt.Errorf("failed to parse github response")
	}

	if repository.Private {
		return fmt.Errorf("private repositories are not supported")
	}

	return nil
}
