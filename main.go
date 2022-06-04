package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	flHelp := flag.Bool("help", false, "display this help message")
	flExm := flag.Bool("example", false, "write example config to stdout")
	flConfig := flag.String("config", "", "path to config")

	flag.Parse()

	if *flHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *flExm {
		fmt.Println(ExampleConfig)
		os.Exit(0)
	}

	if *flConfig == "" {
		flag.Usage()
		fmt.Println("-config must be specified")
		os.Exit(-1)
	}

	c, err := NewConfig(*flConfig)
	if err != nil {
		flag.Usage()
		fmt.Println(err)
		os.Exit(-1)
	}

	mqtt, err := NewMQTT(c.MQTT.Host, c.MQTT.Port, c.MQTT.Username, c.MQTT.Password, c.MQTT.Prefix, c.Devices)
	if err != nil {
		log.Println("could not start mqtt service:", err)
		os.Exit(-1)
	}

	stdout, err := NewCmd(c.RTL433.Path, c.RTL433.ExtraArgs...)
	if err != nil {
		log.Println("could not start rtl_433:", err)
		os.Exit(-1)
	}

	if err = mqtt.Monitor(stdout); err != nil {
		log.Println("error while monitoring rtl_433:", err)
		os.Exit(-1)
	}
}
