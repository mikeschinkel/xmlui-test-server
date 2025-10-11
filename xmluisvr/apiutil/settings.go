package apiutil

import (
	"log"
)

var gitHubRepoURL string

func SetGitHubRepoURL(url string) {
	gitHubRepoURL = url
}

func GitHubRepoURL() string {
	if gitHubRepoURL == "" {
		log.Fatal("GitHub Repo URL not set in apiutil package. Call apiutil.SetGitHubRepoURL() before calling apiutil.GitHubRepoURL().")
	}
	return gitHubRepoURL
}
