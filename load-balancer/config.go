package loadbalancer

import (
	"encoding/json"
	"os"
)

type Backend struct {
	URL    string `json:"url"`
	Health string `json:"health"`
}

type Config struct {
	AlgoType string `json:"algoType"`
	Backends []struct {
		URL    string `json:"url"`
		Health string `json:"health"`
	} `json:"backends"`
}

func ParseConfig() *Config {
	configData, _ := os.ReadFile("./config.json")
	configS := string(configData)
	var config Config
	json.Unmarshal([]byte(configS), &config)
	return &config
}
