// Package storage contains methods for interaction with database.
package storage

import (
	"context"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/model"
)

// HostStorage defines CRUD operations for Host model.
type HostStorage interface {
	GetAll() ([]model.Host, error)
	Get(hostID int) (model.Host, error)
	Save(model.Host) error
	Delete(id int) error
}

// Get eeturns new data service.
func Get(ctx context.Context, appConfig config.Application) (HostStorage, error) {
	return NewYAML(ctx, appConfig.Config.AppHome, appConfig.Logger)
}
