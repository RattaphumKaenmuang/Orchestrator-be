package model

type Instance struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CPU        int    `json:"cpu"`
	RAM        int    `json:"ram"`
	GPU        int    `json:"gpu"`
	Provider   string `json:"provider"`    // "k3s" or "openstack"
	ExternalID string `json:"external_id"` // k3s pod UID or openstack server ID
	Status     string `json:"status"`
}
