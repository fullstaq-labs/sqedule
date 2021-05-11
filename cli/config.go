package cli

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	ServerBaseURL string //`mapstructure:"server-base-url"`
}

func LoadConfigFromViper(viper *viper.Viper) Config {
	return Config{
		ServerBaseURL: viper.GetString("server-base-url"),
	}
}

func (config Config) RequireServerConfig() error {
	if len(config.ServerBaseURL) == 0 {
		return errors.New("Configuration required: server-base-url")
	}

	return nil
}
