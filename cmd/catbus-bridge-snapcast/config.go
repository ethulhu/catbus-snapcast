// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type (
	Config struct {
		MQTTBroker string `json:"mqttBroker"`

		Topics struct {
			Input string `json:"input"`
		} `json:"topics"`

		SnapcastGroupID string `json:"snapcastGroupId"`
	}
)

func loadConfig(path string) (*Config, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	config := &Config{}
	if err := json.Unmarshal(src, config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return config, nil
}
