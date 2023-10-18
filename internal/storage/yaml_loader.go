package storage

import (
	"os"
	"path"

	"github.com/grafviktor/goto/internal/config"
)

func getAppConfigDir(appConfig config.Application) (string, error) {
	// userConfigDir, err := os.Getwd()
	// if err != nil {
	// 	return "", err
	// }
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return path.Join(userConfigDir, appConfig.AppName), nil
}
