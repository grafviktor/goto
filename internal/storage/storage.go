package storage

import (
	"context"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/model"
)

type HostStorage interface {
	GetAll() ([]model.Host, error)
	Get(hostID int) (model.Host, error)
	Save(model.Host) error
	Delete(id int) error
}

func GetStorage(ctx context.Context, config config.Application) (HostStorage, error) {
	return NewYAML(config)
}
