package main

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const TopicBatteryHealthTmpl = "%s/binary_sensor/Govee_%s_battery_health/%s"
const TopicBatteryLevelTmpl = "%s/sensor/Govee_%s_battery_level/%s"
const TopicLastSeen = "%s/sensor/Govee_%s_last_seen/%s"
const TopicMoisture = "%s/binary_sensor/Govee_%s_moisture/%s"
const TopicEndpointConfig = "config"
const TopicEndpointSet = "set"

const QOS2 = byte(2)

type Device struct {
	ID           string `json:"ids"`
	Manufacturer string `json:"mf"`
	Model        string `json:"mdl"`
	Name         string `json:"name"`
}

type Configuration struct {
	Name           string  `json:"name"`
	UniqueID       string  `json:"uniq_id"`
	Device         *Device `json:"dev"`
	DeviceClass    string  `json:"dev_cla,omitempty"`
	EntityCategory string  `json:"ent_cat,omitempty"`
	StateClass     string  `json:"stat_cla,omitempty"`
	StateTopic     string  `json:"stat_t"`
	ValueTemplate  string  `json:"val_tpl,omitempty"`
	Unit           string  `json:"unit_of_meas,omitempty"`
}

type MQTT struct {
	Prefix string
	mqtt.Client
	// devices maps id -> location
	devices map[string]string
}

func NewMQTT(host string, port int, username, password, prefix string, devices map[string]string) (*MQTT, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", host, port)).
		SetUsername(username).
		SetPassword(password)

	client := mqtt.NewClient(opts)
	tok := client.Connect()
	tok.Wait()
	if err := tok.Error(); err != nil {
		return nil, fmt.Errorf("could not connect: %w", err)
	}

	log.Println("mqtt: connected")

	m := &MQTT{Prefix: prefix, Client: client, devices: devices}

	if err := m.ConfigureDevices(); err != nil {
		return nil, fmt.Errorf("could not configure devices: %w", err)
	}

	log.Println("mqtt: devices configured")

	return m, nil
}

func (m *MQTT) Publish(topic string, payload interface{}) mqtt.Token {
	var (
		buf []byte
		err error
	)

	if b, ok := payload.(bool); ok {
		if b {
			buf = []byte("ON")
		} else {
			buf = []byte("OFF")
		}
	} else {
		buf, err = json.Marshal(payload)
		if err != nil {
			panic(fmt.Errorf("unexpected marshal error: %w", err))
		}
	}
	return m.Client.Publish(topic, QOS2, true, buf)
}

func tokenWait(toks ...mqtt.Token) error {
	for _, t := range toks {
		t.Wait()
		if err := t.Error(); err != nil {
			return err
		}
	}
	return nil
}

func (m *MQTT) ConfigureDevice(id string) error {
	location := m.devices[id]

	d := &Device{
		ID:           "Govee_" + id,
		Manufacturer: "Govee",
		Model:        "H5054",
		Name:         location + " Water Sensor",
	}

	toks := make([]mqtt.Token, 0, 4)

	// battery health
	toks = append(toks, m.Publish(fmt.Sprintf(TopicBatteryHealthTmpl, m.Prefix, id, TopicEndpointConfig),
		&Configuration{
			Name:           location + " Water Sensor Battery Health",
			UniqueID:       fmt.Sprintf("Govee_%s_battery_health", id),
			Device:         d,
			DeviceClass:    "battery",
			EntityCategory: "diagnostic",
			StateTopic:     fmt.Sprintf(TopicBatteryHealthTmpl, m.Prefix, id, TopicEndpointSet),
		}))

	// battery level
	toks = append(toks, m.Publish(fmt.Sprintf(TopicBatteryLevelTmpl, m.Prefix, id, TopicEndpointConfig),
		&Configuration{
			Name:           location + " Water Sensor Battery Level",
			UniqueID:       fmt.Sprintf("Govee_%s_battery_level", id),
			Device:         d,
			DeviceClass:    "voltage",
			EntityCategory: "diagnostic",
			StateClass:     "measurement",
			StateTopic:     fmt.Sprintf(TopicBatteryLevelTmpl, m.Prefix, id, TopicEndpointSet),
			Unit:           "V",
		}))

	// last seen
	toks = append(toks, m.Publish(fmt.Sprintf(TopicLastSeen, m.Prefix, id, TopicEndpointConfig),
		&Configuration{
			Name:           location + " Water Sensor Last Seen",
			UniqueID:       fmt.Sprintf("Govee_%s_last_seen", id),
			Device:         d,
			DeviceClass:    "timestamp",
			EntityCategory: "diagnostic",
			StateClass:     "measurement",
			StateTopic:     fmt.Sprintf(TopicLastSeen, m.Prefix, id, TopicEndpointSet),
			ValueTemplate:  "{{ value_json | as_datetime }}",
		}))

	// moisture
	toks = append(toks, m.Publish(fmt.Sprintf(TopicMoisture, m.Prefix, id, TopicEndpointConfig),
		&Configuration{
			Name:        location + " Water Sensor Water Detected",
			UniqueID:    fmt.Sprintf("Govee_%s_moisture", id),
			Device:      d,
			DeviceClass: "moisture",
			StateTopic:  fmt.Sprintf(TopicMoisture, m.Prefix, id, TopicEndpointSet),
		}))

	if err := tokenWait(toks...); err != nil {
		return fmt.Errorf("could not publish configuration: %w", err)
	}

	return nil
}

func (m *MQTT) ConfigureDevices() error {
	for id, location := range m.devices {
		if err := m.ConfigureDevice(id); err != nil {
			return fmt.Errorf("could not configure %s(%s): %w", id, location, err)
		}
	}
	return nil
}

func (m *MQTT) PublishBattery(id string, ok bool, volts float32) error {
	log.Printf("mqtt: battery report from %s (%s): ok: %v, %.2fV\n", id, m.devices[id], ok, volts)

	toks := make([]mqtt.Token, 0, 2)

	// battery health - Home assistant thinks false is good
	toks = append(toks, m.Publish(fmt.Sprintf(TopicBatteryHealthTmpl, m.Prefix, id, TopicEndpointSet), !ok))

	// battery level
	toks = append(toks, m.Publish(fmt.Sprintf(TopicBatteryLevelTmpl, m.Prefix, id, TopicEndpointSet), volts))

	if err := tokenWait(toks...); err != nil {
		return fmt.Errorf("could not publish values: %w", err)
	}

	return nil
}

func (m *MQTT) PublishLastSeen(id string, timestamp string) error {
	tok := m.Publish(fmt.Sprintf(TopicLastSeen, m.Prefix, id, TopicEndpointSet), timestamp)
	tok.Wait()
	if err := tok.Error(); err != nil {
		return fmt.Errorf("could not publish value: %w", err)
	}

	return nil
}

func (m *MQTT) PublishMoisture(id string, wet bool) error {
	if wet {
		log.Printf("mqtt: water leak detected from %s (%s)\n", id, m.devices[id])
	} else {
		log.Printf("mqtt: water leak cleared from %s (%s)\n", id, m.devices[id])
	}

	tok := m.Publish(fmt.Sprintf(TopicMoisture, m.Prefix, id, TopicEndpointSet), wet)
	tok.Wait()
	if err := tok.Error(); err != nil {
		return fmt.Errorf("could not publish value: %w", err)
	}

	return nil
}
