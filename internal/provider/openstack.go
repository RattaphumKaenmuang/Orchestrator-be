package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
)

type OpenStackProvider struct {
	identityClient *gophercloud.ServiceClient
	domainID       string
}

func NewOpenStackProvider() (*OpenStackProvider, error) {
	auth := gophercloud.AuthOptions{
		IdentityEndpoint: os.Getenv("OS_AUTH_URL"),
		Username:         os.Getenv("OS_USERNAME"),
		Password:         os.Getenv("OS_PASSWORD"),
		DomainName:       os.Getenv("OS_USER_DOMAIN_NAME"),
		TenantName:       os.Getenv("OS_PROJECT_NAME"),
	}

	providerClient, err := openstack.AuthenticatedClient(auth)
	if err != nil {
		return nil, fmt.Errorf("openstack auth: %w", err)
	}

	identityClient, err := openstack.NewIdentityV3(providerClient, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("openstack identity: %w", err)
	}

	domainID := os.Getenv("OS_PROJECT_DOMAIN_ID")
	if domainID == "" {
		return nil, fmt.Errorf("OS_PROJECT_DOMAIN_ID is required")
	}

	return &OpenStackProvider{
		identityClient: identityClient,
		domainID:       domainID,
	}, nil
}

func (p *OpenStackProvider) Name() string { return "openstack" }

func (p *OpenStackProvider) CreateGroup(_ context.Context, groupName string) (string, error) {
	project, err := projects.Create(p.identityClient, projects.CreateOpts{
		Name:     groupName,
		DomainID: p.domainID,
		Enabled:  gophercloud.Enabled,
	}).Extract()
	if err != nil {
		return "", err
	}

	return project.ID, nil
}

func (p *OpenStackProvider) DeleteGroup(_ context.Context, groupName string) error {
	allPages, err := projects.List(p.identityClient, projects.ListOpts{
		Name:     groupName,
		DomainID: p.domainID,
	}).AllPages()
	if err != nil {
		return err
	}
	all, err := projects.ExtractProjects(allPages)
	if err != nil {
		return err
	}
	for _, proj := range all {
		if proj.Name == groupName {
			return projects.Delete(p.identityClient, proj.ID).ExtractErr()
		}
	}
	return nil
}

func (p *OpenStackProvider) GroupExists(_ context.Context, groupName string) (bool, error) {
	allPages, err := projects.List(p.identityClient, projects.ListOpts{
		Name:     groupName,
		DomainID: p.domainID,
	}).AllPages()
	if err != nil {
		return false, err
	}
	all, err := projects.ExtractProjects(allPages)
	if err != nil {
		return false, err
	}
	for _, proj := range all {
		if proj.Name == groupName {
			return true, nil
		}
	}
	return false, nil
}
