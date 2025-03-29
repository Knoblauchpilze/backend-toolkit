package config

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
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

	// https://stackoverflow.com/questions/71056755/mapping-string-to-uuid-in-go
	opts := func(decoderConf *mapstructure.DecoderConfig) {
		decoderConf.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			decoderConf.DecodeHook,
			stringToUUIDHookFunc(),
		)
	}

	out := defaultConf
	if err := loader.Unmarshal(&out, opts); err != nil {
		return defaultConf, err
	}

	return out, nil
}
