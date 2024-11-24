package config

import (
	"strings"

	"github.com/spf13/viper"
)

func Load[Configuration any](configName string, defaultConf Configuration) (Configuration, error) {
	loader := viper.New()

	// https://github.com/spf13/viper#reading-config-files
	loader.SetConfigType("yaml")
	loader.AddConfigPath("configs")

	// https://stackoverflow.com/questions/61585304/issues-with-overriding-config-using-env-variables-in-viper
	loader.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	loader.SetEnvPrefix("ENV")
	loader.AutomaticEnv()

	loader.SetConfigName(configName)
	if err := loader.ReadInConfig(); err != nil {
		return defaultConf, err
	}

	out := defaultConf
	if err := loader.Unmarshal(&out); err != nil {
		return defaultConf, err
	}

	return out, nil
}
