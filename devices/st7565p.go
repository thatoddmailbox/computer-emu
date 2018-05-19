package devices

import (
	"sync"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	st7565p_page_width    = 132
	st7565p_page_count    = 8
	st7565p_screen_width  = 128
	st7565p_screen_height = 64
)

// ST7565P display controller, see http://newhavendisplay.com/app_notes/ST7565P.pdf
type ST7565P struct {
	PausedForBreakpoint bool

	columnImmediatelySet bool
	columnAddress        uint8
	pageAddress          uint8
	readModifyWrite      bool

	pio *I8255

	displayMutex  *sync.Mutex
	displayRAM    [st7565p_page_width * (st7565p_page_count + 1)]byte
	displayInvert bool
	displayDirty  bool

	sdlWindow  *sdl.Window
	sdlSurface *sdl.Surface
}

func NewST7565P(pio *I8255) *ST7565P {
	newDevice := &ST7565P{
		displayMutex: &sync.Mutex{},
		pio:          pio,
	}
	sdl.Do(func() {
		var err error
		newDevice.sdlWindow, err = sdl.CreateWindow("Display", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, st7565p_screen_width*2, st7565p_screen_height*2, sdl.WINDOW_OPENGL)
		if err != nil {
			panic(err)
		}
		newDevice.sdlSurface, err = newDevice.sdlWindow.GetSurface()
		if err != nil {
			panic(err)
		}

		go newDevice.drawLoop()
	})
	return newDevice
}

func (d *ST7565P) drawBigPixel(data []byte, x int, y int) {
	d.sdlSurface.FillRect(&sdl.Rect{
		X: int32(x * 2),
		Y: int32(y * 2),
		W: 2,
		H: 2,
	}, 0x0000000)
	// data[4*((2*y*st7565p_screen_width)+2*x)+0] = 0x00
	// data[4*((2*y*st7565p_screen_width)+2*x)+1] = 0x00
	// data[4*((2*y*st7565p_screen_width)+2*x)+2] = 0x00
	// data[4*((2*y*st7565p_screen_width)+2*x)+3] = 0x00
	// data[4*((2*y*st7565p_screen_width)+2*x+1)+0] = 0x00
	// data[4*((2*y*st7565p_screen_width)+2*x+1)+1] = 0x00
	// data[4*((2*y*st7565p_screen_width)+2*x+1)+2] = 0x00
	// data[4*((2*y*st7565p_screen_width)+2*x+1)+3] = 0x00
	// data[4*(((2*y+1)*st7565p_screen_width)+2*x)+0] = 0x00
	// data[4*(((2*y+1)*st7565p_screen_width)+2*x)+1] = 0x00
	// data[4*(((2*y+1)*st7565p_screen_width)+2*x)+2] = 0x00
	// data[4*(((2*y+1)*st7565p_screen_width)+2*x)+3] = 0x00
	// data[4*(((2*y+1)*st7565p_screen_width)+2*x+1)+0] = 0x00
	// data[4*(((2*y+1)*st7565p_screen_width)+2*x+1)+1] = 0x00
	// data[4*(((2*y+1)*st7565p_screen_width)+2*x+1)+2] = 0x00
	// data[4*(((2*y+1)*st7565p_screen_width)+2*x+1)+3] = 0x00
}

func (d *ST7565P) drawLoop() {
	for {
		sdl.Do(func() {
			if d.PausedForBreakpoint {
				return
			}
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch event.(type) {
				case *sdl.KeyboardEvent:
					e := event.(*sdl.KeyboardEvent)
					windowId, err := d.sdlWindow.GetID()
					if err != nil {
						panic(err)
					}
					if e.WindowID != windowId {
						continue
					}
					wasButtonEvent := false
					buttonBit := uint8(0)
					if e.Keysym.Sym == sdl.K_w {
						// up
						wasButtonEvent = true
						buttonBit = 7
					} else if e.Keysym.Sym == sdl.K_s {
						// down
						wasButtonEvent = true
						buttonBit = 6
					} else if e.Keysym.Sym == sdl.K_a {
						// left
						wasButtonEvent = true
						buttonBit = 5
					} else if e.Keysym.Sym == sdl.K_d {
						// right
						wasButtonEvent = true
						buttonBit = 4
					} else if e.Keysym.Sym == sdl.K_f {
						// back
						wasButtonEvent = true
						buttonBit = 3
					} else if e.Keysym.Sym == sdl.K_g {
						// select
						wasButtonEvent = true
						buttonBit = 2
					}
					if wasButtonEvent {
						if e.State == sdl.PRESSED {
							d.pio.SetPortA(d.pio.GetPortA() | (1 << buttonBit))
						} else if e.State == sdl.RELEASED {
							d.pio.SetPortA(d.pio.GetPortA() & ^(1 << buttonBit))
						}
					}
				}
			}
		})

		d.displayMutex.Lock()

		if d.displayDirty {
			d.sdlSurface.FillRect(&sdl.Rect{
				X: 0,
				Y: 0,
				W: d.sdlSurface.W,
				H: d.sdlSurface.H,
			}, 0xFFFFFFFF)

			d.sdlSurface.Lock()

			data := d.sdlSurface.Pixels()

			for y := 0; y < st7565p_page_count; y++ {
				for x := 0; x < st7565p_screen_width; x++ {
					column := d.displayRAM[(y*st7565p_page_width)+x]
					yOffset := 7
					for bit := 0; bit < 8; bit++ {
						bitmask := uint8(1 << uint8(bit))

						value := column & bitmask
						if value != 0 {
							d.drawBigPixel(data, x, (y*8)+yOffset)
						}

						yOffset -= 1
					}
				}
			}

			d.sdlSurface.Unlock()

			d.sdlWindow.UpdateSurface()

			d.displayDirty = false
		}

		d.displayMutex.Unlock()

		time.Sleep(1 / (30 * time.Second))
	}
}

func (d *ST7565P) IsMapped(address uint16) bool {
	if (address&(1<<15) == 0) && (address&(1<<14) != 0) && (address&(1<<13) != 0) && (address&(1<<12) != 0) {
		return true
	}
	return false
}

func (d *ST7565P) incrementColumn() {
	d.columnAddress += 1
	if d.columnAddress >= st7565p_page_width {
		d.columnAddress = 0
	}
}

func (d *ST7565P) ReadByte(address uint16) uint8 {
	maskedAddress := address & 0x800
	if maskedAddress == 0 {
		// status read
		status := uint8(0)
		status |= (1 << 6) // ADC set to normal
		return status
	} else {
		// display data
		if d.columnImmediatelySet {
			// dummy read
			d.columnImmediatelySet = false
			return 0xFF
		}
		d.displayMutex.Lock()
		data := d.displayRAM[(d.pageAddress*st7565p_page_width)+d.columnAddress]
		d.displayMutex.Unlock()
		if !d.readModifyWrite {
			d.incrementColumn()
		}
		return data
	}
}

func (d *ST7565P) WriteByte(address uint16, data uint8) {
	maskedAddress := address & 0x800
	if maskedAddress == 0 {
		// command
		if data == 0xAF {
			// display on
			// panic(errors.New("not implemented"))
		} else if data == 0xAE {
			// display off
			// panic(errors.New("not implemented"))
		} else if data&0xC0 == 0x40 {
			// display start line set
			// panic(errors.New("not implemented"))
		} else if data&0xF0 == 0xB0 {
			// page address set
			d.pageAddress = data & 0xF
		} else if data&0xF0 == 0x1 {
			// column address set high
			value := data & 0xF
			columnLow := d.columnAddress & 0xF
			d.columnAddress = (value << 4) | columnLow
			d.columnImmediatelySet = true
		} else if data&0xF0 == 0x0 {
			// column address set low
			value := data & 0xF
			columnHigh := d.columnAddress & 0xF0
			d.columnAddress = (columnHigh << 4) | value
			d.columnImmediatelySet = true
		} else if data == 0xA0 {
			// adc select normal
			// panic(errors.New("not implemented"))
		} else if data == 0xA1 {
			// adc select reverse
			// panic(errors.New("not implemented"))
		} else if data == 0xA6 {
			// display uninvert
			d.displayMutex.Lock()
			d.displayInvert = false
			d.displayDirty = true
			d.displayMutex.Unlock()
		} else if data == 0xA7 {
			// display invert
			d.displayMutex.Lock()
			d.displayInvert = true
			d.displayDirty = true
			d.displayMutex.Unlock()
		} else if data == 0xA4 {
			// display all points off
			// panic(errors.New("not implemented"))
		} else if data == 0xA5 {
			// display all points on
			// panic(errors.New("not implemented"))
		} else if data == 0xA2 {
			// voltage bias ratio set
			// panic(errors.New("not implemented"))
		} else if data == 0xA3 {
			// voltage bias ratio set
			// panic(errors.New("not implemented"))
		} else if data == 0xE0 {
			// read/modify/write enable
			d.readModifyWrite = true
		} else if data == 0xEE {
			// read/modify/write end
			d.readModifyWrite = false
		} else if data == 0xE2 {
			// reset
			d.columnImmediatelySet = false
			d.columnAddress = 0
			d.pageAddress = 0
			d.readModifyWrite = false

			d.displayMutex.Lock()
			for i := 0; i < len(d.displayRAM); i++ {
				d.displayRAM[i] = 0
			}
			d.displayInvert = false
			d.displayDirty = true
			d.displayMutex.Unlock()
		} else if data&0xC0 == 0xC0 {
			// common output mode select
			// panic(errors.New("not implemented"))
		} else if data&0xF8 == 0x28 {
			// power controller set
			// panic(errors.New("not implemented"))
		} else if data&0xF8 == 0xE3 {
			// nop
		}
	} else {
		// data
		d.displayMutex.Lock()
		d.displayDirty = true
		d.displayRAM[(int(d.pageAddress)*st7565p_page_width)+int(d.columnAddress)] = data
		d.displayMutex.Unlock()
		d.incrementColumn()
	}
}
