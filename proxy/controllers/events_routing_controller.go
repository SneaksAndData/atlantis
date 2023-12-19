package controllers

import (
	"fmt"
	"github.com/google/go-github/v54/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/proxy/services/association"
	"github.com/runatlantis/atlantis/proxy/services/routing"
	"io"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

type EventsRoutingController struct {
	HostAssociation     *association.DefaultHostAssociationService
	EventRouter         *routing.DefaultEventRoutingService
	EventPublishChannel chan *http.Request
	KubernetesClient    *kubernetes.Clientset
}

func (e *EventsRoutingController) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// TODO: add other VCS as well
	go e.routeGithubRequest()
	e.EventPublishChannel <- request
	writer.WriteHeader(200)
}

func (e *EventsRoutingController) parseRequestBody(r *http.Request) ([]byte, error) {
	switch ct := r.Header.Get("Content-Type"); ct {
	case "application/json":
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read body: %s", err)
		}
		return payload, nil
	case "application/x-www-form-urlencoded":
		// GitHub stores the json payload as a form value.
		payloadForm := r.FormValue("payload")
		if payloadForm == "" {
			return nil, errors.New("webhook request did not contain expected 'payload' form value")
		}
		return []byte(payloadForm), nil
	default:
		return nil, fmt.Errorf("webhook request has unsupported Content-Type %q", ct)
	}
}

func (e *EventsRoutingController) routeGithubRequest() {
	for httpRequest := range e.EventPublishChannel {
		payload, err := e.parseRequestBody(httpRequest)
		var prUrl string

		if err != nil {
			// TODO: debug log, go next
			continue
		}

		event, _ := github.ParseWebHook(github.WebHookType(httpRequest), payload)

		switch event := event.(type) {
		case *github.IssueCommentEvent:
			if !event.GetIssue().IsPullRequest() {
				// TODO: debug log, go next
				continue
			}
			prUrl = event.Issue.PullRequestLinks.GetURL()
		case *github.PullRequestEvent:
			prUrl = event.PullRequest.GetURL()
		default:
			// TODO: debug log, go next
			continue
		}

		prAssociation, getErr := (*e.HostAssociation).GetOrReserveHost(prUrl)

		if getErr != nil {
			// TODO: log, go next
			continue
		}

		_, routErr := (*e.EventRouter).RoutePullRequest(*prAssociation, httpRequest)

		if routErr != nil {
			// TODO: log, go next
			continue
		}

		// TODO: print response (info)
	}
}
