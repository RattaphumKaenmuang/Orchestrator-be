package model

type Group struct {
	Name               string     `json:"name"`
	K3sNamespace       string     `json:"k3s_namespace"`
	OpenStackProjectID string     `json:"openstack_project_id"`
	Instances          []Instance `json:"instances"`
}
