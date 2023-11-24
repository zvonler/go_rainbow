package main

import (
	"fmt"
	"image/color"
	"machine"
	"runtime"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const saturation = 255
const brightness = 10

func scale(value, scale uint8) uint8 {
	return uint8((uint16(value) * uint16(scale)) >> 8)
}

func scale_video(value, scale uint8) uint8 {
	var nonzero uint8
	if scale != 0 && scale != 255 {
		nonzero = 1
	}
	return uint8((uint16(value)*uint16(scale))>>8) + nonzero
}

func HSV2RGB(h, s, v uint8) (r, g, b uint8) {
	// Adapted from FastLED:src/hsv2rgb.cpp
	const K255 = 255
	const K171 = 171
	const K170 = 170
	const K85 = 85

	offset := uint8(h & 0x1F)
	offset8 := offset << 3         // offset8 = offset * 8
	third := scale(offset8, 256/3) // max = 85

	if h&0x80 == 0 {
		// 0XX
		if h&0x40 == 0 {
			// 00X
			//section 0-1
			if h&0x20 == 0 {
				// 000
				//case 0: // R -> O
				r, g, b = K255-third, third, 0
			} else {
				// 001
				//case 1: // O -> Y
				r, g, b = K171, K85+third, 0
			}
		} else {
			//01X
			// section 2-3
			if h&0x20 == 0 {
				// 010
				//case 2: // Y -> G
				twothirds := scale(offset8, ((256 * 2) / 3)) // max=170
				r, g, b = K171-twothirds, K170+third, 0
			} else {
				// 011
				// case 3: // G -> A
				r, g, b = 0, K255-third, third
			}
		}
	} else {
		// section 4-7
		// 1XX
		if h&0x40 == 0 {
			// 10X
			if h&0x20 == 0 {
				// 100
				//case 4: // A -> B
				twothirds := scale(offset8, ((256 * 2) / 3)) // max=170
				r, g, b = 0, K171-twothirds, K85+twothirds
			} else {
				// 101
				//case 5: // B -> P
				r, g, b = third, 0, K255-third
			}
		} else {
			if h&0x20 == 0 {
				// 110
				//case 6: // P -- K
				r, g, b = K85+third, 0, K171-third
			} else {
				// 111
				//case 7: // K -> R
				r, g, b = K170+third, 0, K85-third
			}
		}
	}
	return
	// Scale down colors if saturation less than full
	if s != 255 {
		if s == 0 {
			r, g, b = 255, 255, 255
		} else {
			desat := 255 - s
			desat = scale_video(desat, desat)
			satscale := 255 - desat
			if r > 0 {
				r = scale(r, satscale)
			}
			if g > 0 {
				g = scale(g, satscale)
			}
			if b > 0 {
				b = scale(b, satscale)
			}
			r += desat
			g += desat
			b += desat
		}
	}

	// Scale down colors if brightness less than full
	if v != 255 {
		//v = scale_video(v, v)
		if v == 0 {
			r, g, b = 0, 0, 0
		} else {
			r, g, b = scale(r, v)+1, scale(g, v)+1, scale(b, v)+1
		}
	}

	return
}

func fillRainbow(pixels []color.RGBA, offset, step int) {
	for i := 0; i < len(pixels); i++ {
		hue := uint8((offset + step*i) % 256)
		r, g, b := HSV2RGB(hue, saturation, brightness)
		pixels[i] = color.RGBA{r, g, b, 255}

		if i == 0 {
			fmt.Printf("blue:%v, red:%v, green:%v\n", b, r, g)
		}
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
	frameTicker := time.NewTicker(8 * time.Millisecond)

	ledState := false
	toggleHeartbeatLED := func() {
		if ledState = !ledState; ledState {
			led.High()
		} else {
			led.Low()
		}
	}

	updatePixels := func(hue int) {
		fillRainbow(pixelBuf, hue, 25)
		pixels.WriteColors(pixelBuf)
	}

	go func() {
		hue := 0
		for {
			select {
			case <-hueTicker.C:
				hue = (hue + 1) % 256
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
