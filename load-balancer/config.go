package loadbalancer

import (
	"encoding/json"
	"os"
)

type Backend struct {
	URL    string `json:"url"`
	Health string `json:"health"`
	Weight uint   `json:"weight"`
}

type Config struct {
	AlgoType string `json:"algoType"`
	Backends []Backend
}

func ParseConfig() *Config {
	configData, _ := os.ReadFile("./config.json")
	configS := string(configData)
	var config Config
	json.Unmarshal([]byte(configS), &config)
	return &config
}
