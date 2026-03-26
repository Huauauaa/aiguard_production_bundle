package provider

import (
	"net/url"
	"regexp"
	"strings"
)

type RepoInfo struct {
	Provider string `json:"provider"`
	Host     string `json:"host"`
	Owner    string `json:"owner"`
	Name     string `json:"name"`
	Number   string `json:"number"`
	RepoURL  string `json:"repoUrl"`
}

var (
	githubPRPattern = regexp.MustCompile(`^/([^/]+)/([^/]+)/pull/(\d+)$`)
	gitlabMRPattern = regexp.MustCompile(`^/([^/]+)/([^/]+)/-/merge_requests/(\d+)$`)
)

func Parse(raw string) RepoInfo {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RepoInfo{}
	}

	u, err := url.Parse(raw)
	if err != nil {
		return RepoInfo{}
	}

	host := strings.ToLower(u.Host)
	path := strings.TrimSuffix(u.Path, "/")

	if match := githubPRPattern.FindStringSubmatch(path); len(match) == 4 {
		return RepoInfo{
			Provider: "github",
			Host:     host,
			Owner:    match[1],
			Name:     strings.TrimSuffix(match[2], ".git"),
			Number:   match[3],
			RepoURL:  u.Scheme + "://" + host + "/" + match[1] + "/" + strings.TrimSuffix(match[2], ".git") + ".git",
		}
	}

	if match := gitlabMRPattern.FindStringSubmatch(path); len(match) == 4 {
		return RepoInfo{
			Provider: "gitlab",
			Host:     host,
			Owner:    match[1],
			Name:     strings.TrimSuffix(match[2], ".git"),
			Number:   match[3],
			RepoURL:  u.Scheme + "://" + host + "/" + match[1] + "/" + strings.TrimSuffix(match[2], ".git") + ".git",
		}
	}

	return RepoInfo{
		Provider: "generic",
		Host:     host,
	}
}
