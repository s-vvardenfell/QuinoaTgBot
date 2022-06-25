package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func Test_MarshallYmlToConfigStruct(t *testing.T) {
	var cnfg Config

	wd, err := os.Getwd()
	require.NoError(t, err)

	t.Log("\tGetting and marshalling config with viper")
	{
		viper.AddConfigPath(filepath.Join(filepath.Dir(wd), "resources"))
		viper.SetConfigName("config_example")
		viper.SetConfigType("yml")

		err = viper.ReadInConfig()
		require.NoError(t, err)

		err = viper.Unmarshal(&cnfg)
		require.NoError(t, err)
	}
}
