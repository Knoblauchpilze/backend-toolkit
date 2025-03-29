package config

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
)

func stringToUUIDHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to != reflect.TypeOf(uuid.UUID{}) {
			return data, nil
		}

		return uuid.Parse(data.(string))
	}
}
