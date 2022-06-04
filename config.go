package main

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const ExampleConfig = `mqtt:
  host: homeassistant.local
  port: 1883
  username: govee
  password: govee
  prefix: homeassistant
rtl_433:
  path: /usr/local/bin/rtl_433
  extra_args: ["-d", ":1234"]
devices:
  12345: Dishwasher
  54321: Refrigerator`

type Config struct {
	MQTT struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Prefix   string `yaml:"prefix"`
	} `yaml:"mqtt"`
	// Devices is a mapping of ID -> Location
	RTL433 struct {
		Path      string   `yaml:"path"`
		ExtraArgs []string `yaml:"extra_args"`
	} `yaml:"rtl_433"`
	Devices map[string]string `yaml:"devices"`
}

func NewConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config: %w", err)
	}
	defer f.Close()

	c := new(Config)
	if err = yaml.NewDecoder(f).Decode(c); err != nil {
		return nil, fmt.Errorf("could not decode config: %w", err)
	}

	if c.MQTT.Host == "" {
		return nil, errors.New("mqtt.host must be set")
	}

	if c.MQTT.Port == 0 {
		c.MQTT.Port = 1883
	}

	if c.MQTT.Username == "" {
		return nil, errors.New("mqtt.username must be set")
	}

	if c.MQTT.Username == "" {
		return nil, errors.New("mqtt.password must be set")
	}

	if c.MQTT.Prefix == "" {
		c.MQTT.Prefix = "homeassistant"
	}

	if c.RTL433.Path == "" {
		c.RTL433.Path = "rtl_433"
	}

	if len(c.Devices) == 0 {
		return nil, errors.New("devices must be set")
	}

	return c, nil
}
