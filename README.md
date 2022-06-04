# About

hass-mqtt-govee is a small program that wraps around [rtl_433](https://github.com/merbanan/rtl_433) for Govee water leak sensors. It's uses Home Assistant's MQTT auto discovery to automatically add all of your devices.

# Building

    $ cd /path/to/build/directory
    $ GOBIN="$(pwd)" go install "github.com/korylprince/hass-mqtt-govee@<tagged version>"
    $ ./hass-mqtt-govee -h

You must also have a new enough rtl_433 installed that supports protocol 192 (Govee). 

# Configuration

hass-mqtt-govee is configured via a YAML configuration file. Run `./hass-mqtt-govee -example` to output an example file:

```yaml
mqtt:
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
  54321: Refrigerator
```

You can pass extra arguments to rtl_433 with `rtl_433.extra_args`. You can omit this key if you don't need to pass any extra arguments. `devices` should be a mapping from device_id -> location. The easiest way to find the device ID is to run `rtl_433 -R 192`, and press the button on the device. The id will be in the output of the command.

# Using

`./hass-mqtt-govee -config /path/to/config.yaml`

It's recommended to run this with some sort of service scheduler, e.g. systemd, docker, etc.
