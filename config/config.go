// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path"
)

type (
	Config struct {
		BrokerURI string

		Topics struct {
			Input       string
			InputValues string
		}

		Snapcast struct {
			GroupID string
		}
	}

	config struct {
		MQTTBroker string `json:"mqttBroker"`

		Topics struct {
			Input       string `json:"input"`
			InputValues string `json:"inputValues"`
		} `json:"topics"`

		Snapcast struct {
			GroupID string `json:"groupId"`
		} `json:"snapcast"`
	}
)

func ParseFile(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	raw := config{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, err
	}

	return configFromConfig(raw)
}

func configFromConfig(raw config) (*Config, error) {
	c := &Config{
		BrokerURI: raw.MQTTBroker,
	}

	if raw.Topics.Input == "" {
		return nil, errors.New("must set topics.input")
	}
	c.Topics.Input = raw.Topics.Input

	c.Topics.InputValues = raw.Topics.InputValues
	if c.Topics.InputValues == "" {
		c.Topics.InputValues = path.Join(c.Topics.Input, "values")
	}

	if raw.Snapcast.GroupID == "" {
		return nil, errors.New("must set snapcast.groupId")
	}
	c.Snapcast.GroupID = raw.Snapcast.GroupID

	return c, nil
}
