package routing

import (
	"github.com/runatlantis/atlantis/proxy/models"
	"net/http"
)

type EventRoutingService interface {
	routePullRequest(target models.PullRequestAssociation, webhookRequest http.Request) (resp *http.Response, err error)
}

type DefaultEventRoutingService struct {
}

func (ers *DefaultEventRoutingService) routePullRequest(target models.PullRequestAssociation, webhookRequest http.Request) (resp *http.Response, err error) {
	return http.Post(target.AssociatedHost().EventsUrl(), "application/json", webhookRequest.Body)
}
