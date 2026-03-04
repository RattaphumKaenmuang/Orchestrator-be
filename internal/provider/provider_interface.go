package provider

import "context"

type Provider interface {
	Name() string
	CreateGroup(ctx context.Context, groupName string) (string, error)
	DeleteGroup(ctx context.Context, groupName string) error
	GroupExists(ctx context.Context, groupName string) (bool, error)
}
