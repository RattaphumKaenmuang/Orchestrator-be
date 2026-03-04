package provider

import (
	"context"
	"orchestrator/internal/model"
)

type Provider interface {
	Name() string
	CreateGroup(ctx context.Context, groupName string) (string, error)
	DeleteGroup(ctx context.Context, groupName string) error
	GroupExists(ctx context.Context, groupName string) (bool, error)
	CreateInstance(ctx context.Context, group *model.Group, instance *model.Instance) (externalID string, err error)
}
