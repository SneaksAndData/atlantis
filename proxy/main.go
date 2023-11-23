package proxy

import (
	"github.com/runatlantis/atlantis/proxy/controllers"
	"github.com/runatlantis/atlantis/proxy/services/association"
	"github.com/runatlantis/atlantis/proxy/services/routing"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"time"
)

func configureServices() (*kubernetes.Clientset, *association.DefaultHostAssociationService, *routing.DefaultEventRoutingService) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset, association.NewDefaultHostAssociationService(nil, nil), routing.NewEventRoutingService()

}

func main() {
	router := controllers.EventsRoutingController{}
	s := &http.Server{
		Addr:           ":8080",
		Handler:        &router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 10 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
