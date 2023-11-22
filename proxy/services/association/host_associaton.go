package association

import (
	"context"
	"fmt"
	"github.com/runatlantis/atlantis/proxy/models"
	"github.com/runatlantis/atlantis/proxy/services/store"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyconfigurationsautoscalingv1 "k8s.io/client-go/applyconfigurations/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type HostAssociationService interface {
	Exists(prId string) bool
	Unregister(prId string) error
	GetAssociatedHost(prId string) (*models.AssociatedHost, error)
	ReserveHost(prId string, k8s *kubernetes.Clientset, atlantisNamespace string) (*models.PullRequestAssociation, error)
	Consolidate() error
}

type DefaultHostAssociationService struct {
	Store     *store.StateStore[models.PullRequestAssociation]
	HostStore *store.StateStore[models.AssociatedHost]
}

func (d *DefaultHostAssociationService) Exists(prId string) (bool, error) {
	return (*d.Store).Exists(prId)
}

func (d *DefaultHostAssociationService) Unregister(prId string) error {
	return (*d.Store).Remove(prId)
}

func (d *DefaultHostAssociationService) GetAssociatedHost(prId string) (*models.AssociatedHost, error) {
	pra, err := (*d.Store).Get(prId)

	if err != nil {
		return nil, err
	}

	return pra.AssociatedHost(), nil
}

func (d *DefaultHostAssociationService) isReserved(pod v1.Pod) (bool, error) {
	return (*d.HostStore).Exists(pod.Name)
}

func (d *DefaultHostAssociationService) findVacantHost(podList *v1.PodList) (*models.AssociatedHost, error) {
	for _, pod := range podList.Items {
		reserved, err := d.isReserved(pod)
		if err != nil {
			return nil, err
		}

		if !reserved {
			host, hostErr := models.NewAssociatedHost(pod.Name, pod.Status.PodIP)
			if hostErr != nil {
				return nil, hostErr
			}
			return host, nil
		}

		continue
	}

	return nil, nil
}

func (d *DefaultHostAssociationService) waitForNewHost(k8s *kubernetes.Clientset, atlantisNamespace string, hostName string, out chan<- *v1.Pod) {
	timeout := 300      // TODO: from config
	cumulativeWait := 0 // TODO: from config
	step := 10          // TODO: from config
	for {
		pod, err := k8s.CoreV1().Pods(atlantisNamespace).Get(context.TODO(), hostName, metav1.GetOptions{})

		if err != nil {
			out <- nil
		}

		if pod.Status.Phase == v1.PodRunning {
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

func (d *DefaultHostAssociationService) provisionHost(k8s *kubernetes.Clientset, atlantisNamespace string) (*models.AssociatedHost, error) {
	atlantis, err := k8s.AppsV1().StatefulSets(atlantisNamespace).Get(context.TODO(), "atlantis", metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	var newHostNumber = new(int32)
	*newHostNumber = int32(atlantis.Size() + 1)

	scaleConf := applyconfigurationsautoscalingv1.Scale()
	scaleConf.Spec.Replicas = newHostNumber

	_, scaleErr := k8s.AppsV1().StatefulSets(atlantisNamespace).ApplyScale(context.TODO(), "atlantis", scaleConf, metav1.ApplyOptions{})

	if scaleErr != nil {
		return nil, scaleErr
	}

	// wait for scale to complete
	scaleChannel := make(chan *v1.Pod, 1)
	go d.waitForNewHost(k8s, atlantisNamespace, fmt.Sprintf("atlantis-%d", &newHostNumber), scaleChannel)

	newHost := <-scaleChannel

	return models.NewAssociatedHost(newHost.Name, newHost.Status.PodIP)
}

func (d *DefaultHostAssociationService) ReserveHost(prId string, k8s *kubernetes.Clientset, atlantisNamespace string) (*models.PullRequestAssociation, error) {
	// TODO: use https://github.com/enriquebris/goconcurrentqueue for locks
	labelFilters := map[string]string{"app": "atlantis"}
	podList, err := k8s.CoreV1().Pods(atlantisNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: (&(metav1.LabelSelector{MatchLabels: labelFilters})).String()})

	var pra *models.PullRequestAssociation

	if err != nil {
		return nil, err
	}

	vacantHost, searchErr := d.findVacantHost(podList)

	if searchErr != nil {
		return nil, searchErr
	}

	if vacantHost != nil {
		pra = models.NewPullRequestAssociation(prId, *vacantHost)
		(*d.HostStore).Insert(vacantHost.Name(), vacantHost)
	} else {
		newHost, provisionErr := d.provisionHost(k8s, atlantisNamespace)

		if provisionErr != nil {
			return nil, provisionErr
		}

		if newHost != nil {
			(*d.HostStore).Insert(newHost.Name(), newHost)
			pra = models.NewPullRequestAssociation(prId, *newHost)
		}
	}

	return pra, (*d.Store).Insert(prId, pra)
}
