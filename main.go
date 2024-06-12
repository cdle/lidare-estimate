package main

// #cgo LDFLAGS: -lwiringPi -luv
// #include "main.h"
import "C"

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	Keting  = "keting"
	Chufang = "chufang"
	Mov     = "mov"
	Occ     = "occ"
	On      = "on"
	Off     = "off"
)

const dataNum = 200

const occThreshold = 20
const chufangMovThreshold = 50
const ketingMovThreshold = 40

var ketingOccData [dataNum]int
var ketingOccIndex = 0
var ketingOccLast = time.Now()

var chufangOccData [dataNum]int
var chufangOccIndex = 0
var chufangOccLast = time.Now()

//export Receive
func Receive(name *C.char, event *C.char) {
	{
		name := C.GoString(name)
		event := C.GoString(event)
		match := regexp.MustCompile(`(mov|occ), (\d+) (\d+)`).FindStringSubmatch(event)
		if len(match) < 2 {
			return
		}
		tp, _, n2 := match[1], match[2], match[3]
		value, _ := strconv.Atoi(n2)
		// fmt.Println(name, tp, value)
		payload := Off
		switch name {
		case Keting:
			var total = 0
			for _, data := range ketingOccData {
				total += data
			}
			if tp == Occ {
				ketingOccLast = time.Now()
				if total < (value-occThreshold)*dataNum && ketingOccData[ketingOccIndex] != 0 {
					payload = On
				} else {
					ketingOccData[ketingOccIndex] = value
					ketingOccIndex++
					if ketingOccIndex >= dataNum {
						ketingOccIndex %= dataNum
					}
				}
			} else if tp == Mov {
				if total < (value-ketingMovThreshold)*dataNum {
					payload = On
				}
			}
		case Chufang:
			var total = 0
			for _, data := range chufangOccData {
				total += data
			}
			if tp == Occ {
				chufangOccLast = time.Now()
				if total < (value-occThreshold)*dataNum && chufangOccData[chufangOccIndex] != 0 {
					payload = On
				} else {
					chufangOccData[chufangOccIndex] = value
					chufangOccIndex++
					if chufangOccIndex >= dataNum {
						chufangOccIndex %= dataNum
					}
				}
			} else if tp == Mov {
				if total < (value-chufangMovThreshold)*dataNum {
					payload = On
				}
			}
		}
		client.Publish(name+"/"+tp, 0, false, payload)
	}
}

var client mqtt.Client

func main() {
	os.WriteFile("/var/run/uart.pid", []byte(fmt.Sprintf("%d", os.Getpid())), 0755)
	var broker = "192.168.1.1"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetUsername("user")
	opts.SetPassword("password")
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	go func() {
		for {
			time.Sleep(time.Second * 10)
			if ketingOccLast.Add(time.Second * 10).Before(time.Now()) {
				client.Publish("keting"+"/occ", 0, false, Off)
			}
			if chufangOccLast.Add(time.Second * 10).Before(time.Now()) {
				client.Publish("keting"+"/occ", 0, false, Off)
			}
		}
	}()
	C.Start()
}
