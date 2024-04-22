package main

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

func LoadConfig() (PTZInfo, error) {
	info := PTZInfo{}

	dataBytes, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("Failed to load config file.", err)
		return info, err
	}

	err = yaml.Unmarshal(dataBytes, &info)
	if err != nil {
		fmt.Println("Failed to parse config file.", err)
		return info, err
	}

	return info, nil
}