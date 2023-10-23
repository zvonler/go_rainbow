package main

import (
	"image/color"
	"machine"
	"runtime"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

var saturation int = 100
var brightness int = 2

func HSV2RGB(h, s, v int) (int, int, int) {
	var f, p, q, t, r, g, b, hf, sf, vf float64
	var i int
	hf = float64(h) / 360
	sf = float64(s) / 100
	vf = float64(v) / 100

	i = int(hf * 6)
	f = hf*6 - float64(i)
	p = vf * (1 - sf)
	q = vf * (1 - f*sf)
	t = vf * (1 - (1-f)*sf)
	switch i % 6 {
	case 0:
		r = vf
		g = t
		b = p
	case 1:
		r = q
		g = vf
		b = p
	case 2:
		r = p
		g = vf
		b = t
	case 3:
		r = p
		g = q
		b = vf
	case 4:
		r = t
		g = p
		b = vf
	case 5:
		r = vf
		g = p
		b = q
	}
	return int(r * 255), int(g * 255), int(b * 255)
}

func fillRainbow(pixels []color.RGBA, offset, step int) {
	for i := 0; i < len(pixels); i++ {
		hue := (offset + step*i) % 360
		r, g, b := HSV2RGB(hue, saturation, brightness)
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
	frameTicker := time.NewTicker(1 * time.Millisecond)

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
