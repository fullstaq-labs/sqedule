package cli

import (
	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/spf13/viper"
)

func GetViperStringIfSet(viper *viper.Viper, key string) *string {
	if viper.IsSet(key) {
		return lib.NewStringPtr(viper.GetString(key))
	} else {
		return nil
	}
}

func GetViperBoolIfSet(viper *viper.Viper, key string) *bool {
	if viper.IsSet(key) {
		return lib.NewBoolPtr(viper.GetBool(key))
	} else {
		return nil
	}
}
