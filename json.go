package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
)

const EventBatteryReport = "Battery Report"
const EventWaterLeak = "Water Leak"
const EventButtonPress = "Button Press"

const ModelGoveeWater = "Govee-Water"

type RTL433Msg struct {
	Time              string  `json:"time"`
	ID                int     `json:"id"`
	Model             string  `json:"model"`
	Event             string  `json:"event"`
	BatteryOK         float32 `json:"battery_ok"`
	BatteryMilliVolts int     `json:"battery_mV"`
}

func (m *MQTT) Monitor(r io.Reader) error {
	rdr := bufio.NewReader(r)
	for {
		line, err := rdr.ReadBytes('\n')
		if err != nil {
			return fmt.Errorf("could not read line: %w", err)
		}

		msg := new(RTL433Msg)
		if err = json.Unmarshal(line, msg); err != nil {
			log.Printf("monitor: could not json unmarshal line `%s`: %v\n", string(line), err)
			continue
		}

		if msg.Model != ModelGoveeWater {
			log.Printf("monitor: unknown model for %d: `%s`\n", msg.ID, string(line))
			continue
		}

		switch msg.Event {
		case EventBatteryReport:
			if err := m.PublishBattery(strconv.Itoa(msg.ID), msg.BatteryOK == 1, float32(msg.BatteryMilliVolts)/1000); err != nil {
				log.Printf("monitor: could not publish battery status for %d: %v\n", msg.ID, err)
			}
			if err := m.PublishLastSeen(strconv.Itoa(msg.ID), msg.Time); err != nil {
				log.Printf("monitor: could not publish last seen status for %d: %v\n", msg.ID, err)
			}
		case EventWaterLeak:
			if err := m.PublishMoisture(strconv.Itoa(msg.ID), true); err != nil {
				log.Printf("monitor: could not publish moisture status for %d: %v\n", msg.ID, err)
			}
			if err := m.PublishLastSeen(strconv.Itoa(msg.ID), msg.Time); err != nil {
				log.Printf("monitor: could not publish last seen status for %d: %v\n", msg.ID, err)
			}
		case EventButtonPress:
			if err := m.PublishMoisture(strconv.Itoa(msg.ID), false); err != nil {
				log.Printf("monitor: could not publish moisture status for %d: %v\n", msg.ID, err)
			}
			if err := m.PublishLastSeen(strconv.Itoa(msg.ID), msg.Time); err != nil {
				log.Printf("monitor: could not publish last seen status for %d: %v\n", msg.ID, err)
			}
		default:
			log.Printf("monitor: unknown event type for %d: `%s`\n", msg.ID, string(line))
			if err := m.PublishLastSeen(strconv.Itoa(msg.ID), msg.Time); err != nil {
				log.Printf("monitor: could not publish last seen status for %d: %v\n", msg.ID, err)
			}
		}
	}
}
