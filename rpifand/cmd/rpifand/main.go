/*
A daemon using go-rpio library to control the
Raspberry Pi GPIO fan according to the temperature.
Requires administrator rights to run.

Why not using gpiozero?
https://github.com/gpiozero/gpiozero/issues/707
*/

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/logutils"
	rpio "github.com/stianeikeland/go-rpio/v4"
)

var ETC_CONF = "/etc/rpifand/rpifand.toml"
var SYS_TEMP = "/sys/class/thermal/thermal_zone0/temp"
var MODE_ONOFF = "ONOFF"
var MODE_PWM = "PWM"
var Celsius = 1000

type Config struct {
	Main  MainConf  `toml:"MAIN"`
	OnOff OnOffConf `toml:"ONOFF"`
	PWM   PWMConf   `toml:"PWM"`
}

type MainConf struct {
	GPIOPin   int    `toml:"gpio_pin"`
	Mode      string `toml:"mode"`
	Interval  int    `toml:"thermal_interval_seconds"`
	TailRange int    `toml:"tail_range"`
}

type OnOffConf struct {
	CelsiusThreshold int `toml:"temp_threshold_celsius"`
}

type PWMConf struct {
	TempLV0 int    `toml:"temp_celsius_lv0"`
	Speed0  uint32 `toml:"fan_spped_lv0"`
	TempLV1 int    `toml:"temp_celsius_lv1"`
	Speed1  uint32 `toml:"fan_spped_lv1"`
	TempLV2 int    `toml:"temp_celsius_lv2"`
	Speed2  uint32 `toml:"fan_spped_lv2"`
	TempLV3 int    `toml:"temp_celsius_lv3"`
	Speed3  uint32 `toml:"fan_spped_lv3"`
	TempLV4 int    `toml:"temp_celsius_lv4"`
	Speed4  uint32 `toml:"fan_spped_lv4"`
	TempLV5 int    `toml:"temp_celsius_lv5"`
	Speed5  uint32 `toml:"fan_spped_lv5"`
}

func loadConf() (*Config, error) {
	var conf Config
	log.Printf("[I] rpifand: load config from %s\n", ETC_CONF)
	if _, err := toml.DecodeFile(ETC_CONF, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

type FanD struct {
	c        chan int
	exitChan chan os.Signal
	conf     *Config
	pin      rpio.Pin
}

func NewFand() (*FanD, error) {
	if err := rpio.Open(); err != nil {
		return nil, err
	}

	conf, err := loadConf()
	if err != nil {
		return nil, err
	}

	c := make(chan int)
	exitChan := make(chan os.Signal)
	pin := rpio.Pin(conf.Main.GPIOPin)
	return &FanD{c, exitChan, conf, pin}, nil
}

func (fand *FanD) Run() {
	go fand.temPoll()
	go fand.gpioFan()
	fand.sigWatch()
}

func (fand *FanD) sigWatch() {
	// SIGINT: Ctrl + C --> interrupt
	// SIGTERM: man systemd.kill for stop and restart service
	signal.Notify(fand.exitChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("[D] rpifand: wait for signal")

	sig := <-fand.exitChan
	log.Printf("[I] rpifand: received signal: %v.", sig)
	log.Printf("[I] rpifand: reset gpio fan and exit.")

	fand.pin.Output()
	fand.pin.High()
	rpio.Close()

	log.Println("[I] rpifand: bye!")
	os.Exit(0)
}

func (fand *FanD) temPoll() {
	tick := time.Tick(time.Duration(fand.conf.Main.Interval) * time.Second)

	for {
		data, err := os.ReadFile(SYS_TEMP)
		if err != nil {
			log.Printf("[E] rpifand: read temperature failed: %v", err)
			fand.c <- 100 * Celsius
			continue
		}

		dataStr := strings.TrimSuffix(string(data), "\n")
		temp, err := strconv.Atoi(dataStr)
		if err != nil {
			log.Printf("[E] rpifand: the content read cannot be converted to an integer: %v", err)
			fand.c <- 100 * Celsius
			continue
		}

		fand.c <- temp

		<-tick
	}
}

func (fand *FanD) gpioFan() {
	if fand.conf.Main.Mode == MODE_ONOFF {
		log.Printf("[I] rpifand: enter ON/OFF mode loop, temperature threshold setting: %d Celsius.\n", fand.conf.OnOff.CelsiusThreshold)
		fand.onoffLoop()
	} else {
		log.Printf("[I] rpifand: enter PWM mode loop, temperature range setting: %d - %d Celsius.\n", fand.conf.PWM.TempLV0, fand.conf.PWM.TempLV5)
		fand.pwmLoop()
	}
}

func (fand *FanD) onoffLoop() {
	fand.pin.Output()
	fand.pin.High()

	lastStatus := rpio.High
	tempRecords := make([]int, fand.conf.Main.TailRange)
	for {
		temp := <-fand.c
		tempRecords = append(tempRecords, temp)
		log.Printf("[D] rpifand: current temperature record: %v", tempRecords)

		statxt := ""
		status := lastStatus
		tempRecords = tempRecords[1:]
		maxTemp := MaxIntSlice(tempRecords)
		if maxTemp >= fand.conf.OnOff.CelsiusThreshold*Celsius {
			status = rpio.High
			statxt = "ON"
		} else {
			status = rpio.Low
			statxt = "OFF"
		}
		if status != lastStatus {
			log.Printf("[I] rpifand: toggle gpio fan to %v for temperature %d: %v", statxt, maxTemp/Celsius, tempRecords)
			rpio.WritePin(fand.pin, status)
			lastStatus = status
		}
	}
}

func (fand *FanD) pwmLoop() {
	fand.pin.Pwm()
	// pull up gpio fan, keep 100% running
	fand.pin.DutyCycle(100, 100)
	fand.pin.Freq(38000 * 4)

	lastRate := uint32(100)
	tempRecords := make([]int, fand.conf.Main.TailRange)
	for {
		temp := <-fand.c
		tempRecords = append(tempRecords, temp)
		log.Printf("[D] rpifand: current temperature record: %v", tempRecords)

		rate := lastRate
		tempRecords = tempRecords[1:]
		maxTemp := MaxIntSlice(tempRecords)
		if maxTemp < fand.conf.PWM.TempLV0*Celsius {
			rate = fand.conf.PWM.Speed0
		} else if (fand.conf.PWM.TempLV1*Celsius <= maxTemp) && (maxTemp < fand.conf.PWM.TempLV2*Celsius) {
			rate = fand.conf.PWM.Speed1
		} else if (fand.conf.PWM.TempLV2*Celsius <= maxTemp) && (maxTemp < fand.conf.PWM.TempLV3*Celsius) {
			rate = fand.conf.PWM.Speed2
		} else if (fand.conf.PWM.TempLV3*Celsius <= maxTemp) && (maxTemp < fand.conf.PWM.TempLV4*Celsius) {
			rate = fand.conf.PWM.Speed3
		} else if (fand.conf.PWM.TempLV4*Celsius <= maxTemp) && (maxTemp < fand.conf.PWM.TempLV5*Celsius) {
			rate = fand.conf.PWM.Speed4
		} else {
			rate = fand.conf.PWM.Speed5
		}
		if rate != lastRate {
			log.Printf("[I] rpifand: change PWM duty-cycle to %d/100 for temperature %d: %v", rate, maxTemp/Celsius, tempRecords)
			fand.pin.DutyCycle(rate, 100)
			lastRate = rate
		}
	}
}

func MaxIntSlice(v []int) int {
	if len(v) == 0 {
		return 0
	}

	m := v[0]
	for i := 1; i < len(v); i++ {
		if v[i] > m {
			m = v[i]
		}
	}
	return m
}

var logLevel = "INFO"

func main() {
	flag.StringVar(&logLevel, "log_level", logLevel, "log level")
	flag.Parse()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"D", "I", "W", "E"},
		MinLevel: logutils.LogLevel(strings.ToUpper(logLevel)[:1]),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	fand, err := NewFand()
	if err != nil {
		log.Fatalf("[E] rpifand: exit due to error: %v\n", err)
	}
	fand.Run()
}
