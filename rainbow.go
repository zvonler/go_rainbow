package main

import (
	"image/color"
	"machine"
	"runtime"
	"time"

	cc "github.com/SimonWaldherr/ColorConverterGo"
	"tinygo.org/x/drivers/ws2812"
)

var saturation int = 100
var brightness int = 2

func fillRainbow(pixels []color.RGBA, offset, step int) {
	for i := 0; i < len(pixels); i++ {
		hue := (offset + step*i) % 360
		r, g, b := cc.HSV2RGB(hue, saturation, brightness)
		pixels[i] = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	}
}

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.NEOPIXELS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pixels := ws2812.New(machine.NEOPIXELS)

	pixelBuf := make([]color.RGBA, 10)
	pixels.WriteColors(pixelBuf)

	heartbeatTicker := time.NewTicker(512 * time.Millisecond)
	hueTicker := time.NewTicker(40 * time.Millisecond)
	frameTicker := time.NewTicker(30 * time.Millisecond)

	ledState := false
	toggleHeartbeatLED := func() {
		if ledState = !ledState; ledState {
			led.High()
		} else {
			led.Low()
		}
	}

	updatePixels := func(hue int) {
		fillRainbow(pixelBuf, hue, 36)
		pixels.WriteColors(pixelBuf)
	}

	go func() {
		hue := 0
		for {
			select {
			case <-hueTicker.C:
				hue = (hue + 1) % 360
			case <-heartbeatTicker.C:
				toggleHeartbeatLED()
			case <-frameTicker.C:
				updatePixels(hue)
			default:
				runtime.Gosched()
			}
		}
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}
