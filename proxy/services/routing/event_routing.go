package routing

import (
	"github.com/runatlantis/atlantis/proxy/models"
	"net/http"
)

type EventRoutingService interface {
	routePullRequest(target models.PullRequestAssociation, webhookRequest http.Request) (resp *http.Response, err error)
}

func routePullRequest(ers *EventRoutingService, target models.PullRequestAssociation, webhookRequest http.Request) (resp *http.Response, err error) {
	const atlantisEventsControllerPath = "/events"

	return http.Post(target.AssociatedHost.EventsUrl(), "application/json", webhookRequest.Body)
}
