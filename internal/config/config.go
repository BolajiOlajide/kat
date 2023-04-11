package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/BolajiOlajide/kat/internal/constants"
	"github.com/BolajiOlajide/kat/internal/types"
)

var errConfigNotFound = errors.Newf("config file not provided. Provide one via the `-c` flag or have the %s in the current working directory.", constants.KatConfigurationFileName)

func GetKatConfigFromCtx(c *cli.Context) (types.Config, error) {
	cfg, ok := c.Context.Value(constants.KatConfigKey).(types.Config)
	if !ok {
		return types.Config{}, errors.New("invalid configuration")
	}
	return cfg, nil
}

func ParseConfig(c *cli.Context) error {
	var (
		f   []byte
		err error
	)

	configPath := c.String("config")
	if !filepath.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "getting working directory")
		}
		configPath = filepath.Join(wd, configPath)
	}

	fmt.Println(configPath, "configuration path")
	f, err = os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errConfigNotFound
		}
		return err
	}

	var cfg = &types.Config{}
	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return errors.Wrap(err, "marshalling config")
	}

	cfg.SetDefault()
	newContext := context.WithValue(c.Context, constants.KatConfigKey, *cfg)
	c.Context = newContext
	return nil
}
