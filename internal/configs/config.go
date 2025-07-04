package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Access struct {
	Name string
	Key  string
}

type Resource struct {
	Name            string
	Endpoint        string
	Destination_URL string
	Access          []Access
}

type Server struct {
	Host        string
	Listen_port string
}

type Configuration struct {
	Server    Server
	Resources []Resource
}

var Config *Configuration

func NewConfiguration() (*Configuration, error) {
	viper.AddConfigPath("settings")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %s", err)
	}
	err = viper.Unmarshal(&Config)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	return Config, nil
}
