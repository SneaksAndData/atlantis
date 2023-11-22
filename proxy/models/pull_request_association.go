package models

import "time"

type PullRequestAssociation struct {
	PullRequestId  string
	AssociatedHost AssociatedHost
	CreatedAt      time.Time
}
