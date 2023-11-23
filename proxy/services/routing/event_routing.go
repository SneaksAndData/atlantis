package routing

import (
	"github.com/runatlantis/atlantis/proxy/models"
	"net/http"
)

type EventRoutingService interface {
	RoutePullRequest(target models.PullRequestAssociation, webhookRequest *http.Request) (resp *http.Response, err error)
}

type DefaultEventRoutingService struct {
}

func NewEventRoutingService() *DefaultEventRoutingService {
	return &DefaultEventRoutingService{}
}

func (ers *DefaultEventRoutingService) RoutePullRequest(target models.PullRequestAssociation, webhookRequest *http.Request) (resp *http.Response, err error) {
	return http.Post(target.AssociatedHost().EventsUrl(), "application/json", webhookRequest.Body)
}
