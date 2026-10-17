package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePathIfExist() (string, error) {
	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	filePath := homeDirPath + "/" + configFileName

	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("[Error] Couldn't find the config file on path <%s>: %w", filePath, err)
	}

	return filePath, err
}

func Read() (Config, error) {
	filePath, err := getConfigFilePathIfExist()
	if err != nil {
		return Config{}, err
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func writeConfigFile(config *Config) error {
	filePath, err := getConfigFilePathIfExist()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, bytes, 0o644)
}

func (config *Config) SetUser(CurrentUserame string) error {
	config.CurrentUsername = CurrentUserame

	return writeConfigFile(config)
}
