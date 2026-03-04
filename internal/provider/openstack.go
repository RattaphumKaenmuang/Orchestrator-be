package provider

import (
	"context"
	"fmt"
	"os"

	"orchestrator/internal/model"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/roles"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"
)

type OpenStackProvider struct {
	providerClient *gophercloud.ProviderClient
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

	identityClient, err := openstack.NewIdentityV3(providerClient, gophercloud.EndpointOpts{
		Availability: gophercloud.AvailabilityPublic,
	})
	if err != nil {
		return nil, fmt.Errorf("openstack identity: %w", err)
	}

	domainID := os.Getenv("OS_PROJECT_DOMAIN_ID")
	if domainID == "" {
		return nil, fmt.Errorf("OS_PROJECT_DOMAIN_ID is required")
	}

	return &OpenStackProvider{
		providerClient: providerClient,
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

	// Assign admin role to admin user so it shows up in Horizon
	adminUserID, err := p.findUserID(os.Getenv("OS_USERNAME"))
	if err != nil {
		return "", fmt.Errorf("find admin user: %w", err)
	}

	adminRoleID, err := p.findRoleID("admin")
	if err != nil {
		return "", fmt.Errorf("find admin role: %w", err)
	}

	err = roles.Assign(p.identityClient, adminRoleID, roles.AssignOpts{
		UserID:    adminUserID,
		ProjectID: project.ID,
	}).ExtractErr()
	if err != nil {
		return "", fmt.Errorf("assign role: %w", err)
	}

	return project.ID, nil
}

func (p *OpenStackProvider) findUserID(username string) (string, error) {
	allPages, err := users.List(p.identityClient, users.ListOpts{
		Name: username,
	}).AllPages()
	if err != nil {
		return "", err
	}
	allUsers, err := users.ExtractUsers(allPages)
	if err != nil {
		return "", err
	}
	for _, u := range allUsers {
		if u.Name == username {
			return u.ID, nil
		}
	}
	return "", fmt.Errorf("user %q not found", username)
}

func (p *OpenStackProvider) findRoleID(roleName string) (string, error) {
	allPages, err := roles.List(p.identityClient, roles.ListOpts{
		Name: roleName,
	}).AllPages()
	if err != nil {
		return "", err
	}
	allRoles, err := roles.ExtractRoles(allPages)
	if err != nil {
		return "", err
	}
	for _, r := range allRoles {
		if r.Name == roleName {
			return r.ID, nil
		}
	}
	return "", fmt.Errorf("role %q not found", roleName)
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

func (p *OpenStackProvider) CreateInstance(_ context.Context, group *model.Group, instance *model.Instance) (string, error) {
	auth := gophercloud.AuthOptions{
		IdentityEndpoint: os.Getenv("OS_AUTH_URL"),
		Username:         os.Getenv("OS_USERNAME"),
		Password:         os.Getenv("OS_PASSWORD"),
		DomainName:       os.Getenv("OS_USER_DOMAIN_NAME"),
		TenantID:         group.OpenStackProjectID,
	}

	scopedClient, err := openstack.AuthenticatedClient(auth)
	if err != nil {
		return "", fmt.Errorf("openstack scoped auth: %w", err)
	}

	computeClient, err := openstack.NewComputeV2(scopedClient, gophercloud.EndpointOpts{
		Availability: gophercloud.AvailabilityPublic,
	})
	if err != nil {
		return "", fmt.Errorf("openstack compute client: %w", err)
	}

	flavorID, err := p.findFlavor(computeClient, instance.CPU, instance.RAM)
	if err != nil {
		return "", err
	}

	imageID := os.Getenv("OS_DEFAULT_IMAGE_ID")
	if imageID == "" {
		return "", fmt.Errorf("OS_DEFAULT_IMAGE_ID env var is required")
	}

	networkID := os.Getenv("OS_DEFAULT_NETWORK_ID")
	if networkID == "" {
		return "", fmt.Errorf("OS_DEFAULT_NETWORK_ID env var is required")
	}

	server, err := servers.Create(computeClient, servers.CreateOpts{
		Name:      instance.Name,
		FlavorRef: flavorID,
		ImageRef:  imageID,
		Networks: []servers.Network{
			{UUID: networkID},
		},
		Metadata: map[string]string{
			"orchestrator_group": group.Name,
		},
	}).Extract()
	if err != nil {
		return "", fmt.Errorf("openstack create server: %w", err)
	}

	return server.ID, nil
}

func (p *OpenStackProvider) findFlavor(computeClient *gophercloud.ServiceClient, cpu, ramMB int) (string, error) {
	allPages, err := flavors.ListDetail(computeClient, flavors.ListOpts{}).AllPages()
	if err != nil {
		return "", fmt.Errorf("list flavors: %w", err)
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return "", fmt.Errorf("extract flavors: %w", err)
	}

	var bestID string
	bestVCPUs := int(^uint(0) >> 1) // max int
	bestRAM := int(^uint(0) >> 1)

	for _, f := range allFlavors {
		if f.VCPUs >= cpu && f.RAM >= ramMB {
			if f.VCPUs < bestVCPUs || (f.VCPUs == bestVCPUs && f.RAM < bestRAM) {
				bestID = f.ID
				bestVCPUs = f.VCPUs
				bestRAM = f.RAM
			}
		}
	}

	if bestID == "" {
		return "", fmt.Errorf("no flavor found with >= %d vCPUs and >= %d MB RAM", cpu, ramMB)
	}
	return bestID, nil
}
