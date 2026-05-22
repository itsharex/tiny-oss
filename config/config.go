package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Conf struct {
	DB struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Security struct {
		Token    string `yaml:"token"`
		APIToken string `yaml:"api_token"`
	} `yaml:"security"`
	App struct {
		UploadDir   string `yaml:"upload_dir"`
		FileBaseURL string `yaml:"file_base_url"`
	} `yaml:"app"`
}

var Config Conf

func InitConfig() error {
	file, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer file.Close()
	return yaml.NewDecoder(file).Decode(&Config)
}
