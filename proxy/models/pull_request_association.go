package models

import (
	"github.com/google/uuid"
	"time"
)

type PullRequestAssociation struct {
	pullRequestId  string
	pullRequestUrl string
	associatedHost AssociatedHost
	createdAt      time.Time
}

func NewPullRequestAssociation(prUrl string, host AssociatedHost) *PullRequestAssociation {
	return &PullRequestAssociation{
		pullRequestId:  uuid.New().String(),
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
