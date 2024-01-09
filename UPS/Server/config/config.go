// package config

// import (
// 	"fmt"

// 	"github.com/spf13/viper"
// )

// type Configuration struct {
// 	Server struct {
// 		Host string
// 		Port string
// 	}
// }

// func LoadConfiguration(configPath, configName, ConfigType string) (*Configuration, error) {
// 	var config *Configuration

// 	viper.AddConfigPath(configPath)
// 	viper.SetConfigName(configName)
// 	viper.SetConfigType(ConfigType)

// 	err := viper.ReadInConfig()
// 	if err != nil {
// 		return nil, fmt.Errorf("Could not read config: %v", err)
// 	}

// 	err = viper.Unmarshal(&config)
// 	if err != nil {
// 		return nil, fmt.Errorf("Could not unmarshal: %v", err)
// 	}

// 	return config, nil
// }
