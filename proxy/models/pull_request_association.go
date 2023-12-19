package models

import (
	"net/url"
	"strings"
	"time"
)

type PullRequestAssociation struct {
	pullRequestId  string
	pullRequestUrl string
	associatedHost AssociatedHost
	createdAt      time.Time
}

func NewPullRequestAssociation(prUrl string, host AssociatedHost) *PullRequestAssociation {
	parsed, err := url.Parse(prUrl)

	if err != nil {
		return nil
	}

	return &PullRequestAssociation{
		pullRequestId:  strings.ToLower(strings.Replace(parsed.Path, "/", "", -1)),
		pullRequestUrl: prUrl,
		associatedHost: host,
		createdAt:      time.Now().UTC(),
	}
}

func (pra *PullRequestAssociation) PullRequestUrl() string {
	return pra.pullRequestUrl
}

func (pra *PullRequestAssociation) PullRequestId() string {
	return pra.pullRequestId
}

func (pra *PullRequestAssociation) AssociatedHost() *AssociatedHost {
	return &pra.associatedHost
}

func (pra *PullRequestAssociation) CreatedAt() time.Time {
	return pra.createdAt
}
