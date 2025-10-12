package apiresp

import (
	"log"
)

var gitHubRepoURL string

func SetGitHubRepoURL(url string) {
	gitHubRepoURL = url
}

func GitHubRepoURL() string {
	if gitHubRepoURL == "" {
		log.Fatal("GitHub Repo URL not set in apiresp package. Call apiresp.SetGitHubRepoURL() before calling apiresp.GitHubRepoURL().")
	}
	return gitHubRepoURL
}
