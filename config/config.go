// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// Config holds all configuration needed by this app.
type Config struct {
	MariaDbUser     string `envconfig:"DB_HR_USER" required:"true"`
	MariaDbPassword string `envconfig:"DB_PASSWORD" required:"true"`
	MariaDbDatabase string `envconfig:"DB_SCHEMA" required:"true"`
	MariaDbHost     string `envconfig:"DB_HOST_NAME" required:"true"`
	MariaDbPort     string `envconfig:"DB_PORT" required:"true"`
}

var (
	godotenvLoad     = godotenv.Load
	envconfigProcess = envconfig.Process
)

func ReadConfig() (*Config, error) {
	if err := godotenvLoad(); err != nil {
		return nil, errors.Wrap(err, "loading env vars")
	}
	config := new(Config)
	if err := envconfigProcess("", config); err != nil {
		return nil, errors.Wrap(err, "processing env vars")
	}
	return config, nil
}
