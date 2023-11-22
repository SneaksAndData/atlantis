package association

import "github.com/runatlantis/atlantis/proxy/models"

type HostAssociationService interface {
	exists(prId string) bool
	register(prId string, host models.AssociatedHost)
	unregister(prId string)
}
