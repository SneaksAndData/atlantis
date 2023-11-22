package models

import (
	"fmt"
	"net"
)

type AssociatedHost struct {
	hostName string
	hostIP   net.IP
}

func (h *AssociatedHost) EventsUrl() string {
	return fmt.Sprintf("http://%s/events", h.hostIP.String())
}
