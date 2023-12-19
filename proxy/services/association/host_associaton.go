package association

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/proxy/models"
	v1batch "k8s.io/api/batch/v1"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type HostAssociationService interface {
	Unregister(prId string) error
	GetOrReserveHost(prId string, k8s *kubernetes.Clientset, atlantisNamespace string) (*models.PullRequestAssociation, error)
	Consolidate() error
}

type DefaultHostAssociationService struct {
	atlantisNamespace string
	k8s               *kubernetes.Clientset
	jobTemplate       *v1batch.Job
}

func NewDefaultHostAssociationService(kubernetesClient *kubernetes.Clientset) *DefaultHostAssociationService {
	return &DefaultHostAssociationService{atlantisNamespace: "atlantis", k8s: kubernetesClient}
}

func (d *DefaultHostAssociationService) waitForNewHost(hostName string, out chan<- *v1core.Pod) {
	timeout := 300      // TODO: from config
	cumulativeWait := 0 // TODO: from config
	step := 10          // TODO: from config
	for {
		pod, err := d.k8s.CoreV1().Pods(d.atlantisNamespace).Get(context.TODO(), hostName, metav1.GetOptions{})

		if err != nil {
			out <- nil
		}

		if pod.Status.Phase == v1core.PodRunning {
			out <- pod
			break
		}

		if cumulativeWait >= timeout {
			out <- nil
			break
		}

		time.Sleep(time.Duration(step) * time.Second)
		cumulativeWait = cumulativeWait + step
		continue
	}
}

func (d *DefaultHostAssociationService) provisionHost(prId string) (*models.AssociatedHost, error) {
	newJob := d.jobTemplate
	newJob.SetName(prId)
	newJob.SetLabels(map[string]string{"app": "atlantis-job", "prId": prId})

	_, err := d.k8s.BatchV1().Jobs(d.atlantisNamespace).Create(context.TODO(), newJob, metav1.CreateOptions{})

	if err != nil {
		return nil, err
	}

	newAtlantisPod, searchErr := d.getHost(prId)

	if searchErr != nil {
		return nil, searchErr
	}

	// wait for scale to complete
	scaleChannel := make(chan *v1core.Pod, 1)
	go d.waitForNewHost(newAtlantisPod.Name(), scaleChannel)

	newHost := <-scaleChannel

	return models.NewAssociatedHost(newHost.Name, newHost.Status.PodIP)
}

func (d *DefaultHostAssociationService) getHost(prId string) (*models.AssociatedHost, error) {
	job, err := d.k8s.BatchV1().Jobs(d.atlantisNamespace).Get(context.TODO(), prId, metav1.GetOptions{})
	labelFilters := map[string]string{"app": "atlantis-job", "prId": prId}

	if err != nil {
		return nil, err
	}

	if job != nil {
		podList, listErr := d.k8s.CoreV1().Pods(d.atlantisNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: (&(metav1.LabelSelector{MatchLabels: labelFilters})).String()})

		if listErr != nil {
			return nil, listErr
		}

		if len(podList.Items) == 1 {
			return models.NewAssociatedHost(podList.Items[0].Name, podList.Items[0].Status.PodIP)
		}

		return nil, errors.New(fmt.Sprintf("More than a single pod match PR %s", prId))
	}

	return nil, nil
}

func (d *DefaultHostAssociationService) GetOrReserveHost(prId string) (*models.PullRequestAssociation, error) {
	maybeHost, getHostErr := d.getHost(prId)

	var pra *models.PullRequestAssociation

	if getHostErr != nil {
		return nil, getHostErr
	}

	if maybeHost != nil {
		pra = models.NewPullRequestAssociation(prId, *maybeHost)
	} else {
		newHost, provisionErr := d.provisionHost(prId)

		if provisionErr != nil {
			return nil, provisionErr
		}

		pra = models.NewPullRequestAssociation(prId, *newHost)
	}

	return pra, nil
}
