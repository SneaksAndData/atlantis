package models

import (
	"errors"
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

func (h *AssociatedHost) Name() string {
	return h.hostName
}

func NewAssociatedHost(hostName string, hostIP string) (*AssociatedHost, error) {
	var ip = net.ParseIP(hostIP)

	if ip == nil {
		return nil, errors.New(fmt.Sprintf("Invalid host IP address: %s", hostIP))
	}

	return &AssociatedHost{hostName: hostName, hostIP: ip}, nil
}
