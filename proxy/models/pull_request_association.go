package models

import (
	"time"
)

type PullRequestAssociation struct {
	pullRequestId  string
	associatedHost AssociatedHost
	createdAt      time.Time
}

func NewPullRequestAssociation(prId string, host AssociatedHost) *PullRequestAssociation {
	return &PullRequestAssociation{
		pullRequestId:  prId,
		associatedHost: host,
		createdAt:      time.Now().UTC(),
	}
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
