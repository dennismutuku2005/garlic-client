package garlic

import (
	"fmt"
)

// SystemResource represents the hardware and software resource details of the router.
type SystemResource struct {
	Uptime           string
	Version          string
	CPU              string
	CPUCount         string
	CPUFrequency     string
	CPULoad          string
	FreeMemory       string
	TotalMemory      string
	FreeHDDSpace     string
	TotalHDDSpace    string
	ArchitectureName string
	BoardName        string
	Platform         string
}

// SystemIdentity represents the router identity (hostname/name).
type SystemIdentity struct {
	Name string
}

// Interface represents a network interface configured on the router.
type Interface struct {
	ID        string
	Name      string
	Type      string
	MTU       string
	ActualMTU string
	Running   bool
	Disabled  bool
}

// SystemResource fetches the resource specifications and usage from the router.
func (c *Client) SystemResource() (*SystemResource, error) {
	reply, err := c.RunOne(NewCommand("/system/resource/print"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch system resources: %w", err)
	}

	res := &SystemResource{
		Uptime:           reply.Map["uptime"],
		Version:          reply.Map["version"],
		CPU:              reply.Map["cpu"],
		CPUCount:         reply.Map["cpu-count"],
		CPUFrequency:     reply.Map["cpu-frequency"],
		CPULoad:          reply.Map["cpu-load"],
		FreeMemory:       reply.Map["free-memory"],
		TotalMemory:      reply.Map["total-memory"],
		FreeHDDSpace:     reply.Map["free-hdd-space"],
		TotalHDDSpace:    reply.Map["total-hdd-space"],
		ArchitectureName: reply.Map["architecture-name"],
		BoardName:        reply.Map["board-name"],
		Platform:         reply.Map["platform"],
	}
	return res, nil
}

// SystemIdentity retrieves the router's current identity.
func (c *Client) SystemIdentity() (*SystemIdentity, error) {
	reply, err := c.RunOne(NewCommand("/system/identity/print"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch system identity: %w", err)
	}

	return &SystemIdentity{
		Name: reply.Map["name"],
	}, nil
}

// InterfaceList retrieves a list of all network interfaces configured on the router.
func (c *Client) InterfaceList() ([]*Interface, error) {
	replies, err := c.Run(NewCommand("/interface/print"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch interfaces: %w", err)
	}

	var interfaces []*Interface
	for _, reply := range replies {
		if reply.Type != "!re" {
			continue
		}

		ifaces := &Interface{
			ID:        reply.Map[".id"],
			Name:      reply.Map["name"],
			Type:      reply.Map["type"],
			MTU:       reply.Map["mtu"],
			ActualMTU: reply.Map["actual-mtu"],
			Running:   reply.Map["running"] == "true",
			Disabled:  reply.Map["disabled"] == "true",
		}
		interfaces = append(interfaces, ifaces)
	}
	return interfaces, nil
}
