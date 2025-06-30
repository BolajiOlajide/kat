// Package config handles loading and managing configuration for the kat migration tool.
// It supports loading configuration from YAML files and provides utilities for
// managing configuration context within CLI commands.
//
// The package supports:
//   - Loading configuration from YAML files
//   - Environment variable substitution in configuration
//   - Context-based configuration management for CLI commands
//   - Validation of configuration parameters
package config

import (
	"context"
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

	f, err = os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errConfigNotFound
		}
		return err
	}

	// Expand environment variables in the config
	configStr := os.ExpandEnv(string(f))

	var cfg = &types.Config{}
	err = yaml.Unmarshal([]byte(configStr), cfg)
	if err != nil {
		return errors.Wrap(err, "marshalling config")
	}

	cfg.SetDefault()
	c.Context = context.WithValue(c.Context, constants.KatConfigKey, *cfg)
	return nil
}
