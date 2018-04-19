package debugger

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/thatoddmailbox/minemu/cpu"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type DebuggerState struct {
	CPU cpu.CPU
}

type Debugger struct {
	StepChannel chan bool
	CPU         *cpu.CPU
	CPUMutex    *sync.Mutex
	SingleStep  bool
}

func NewDebugger(sim *cpu.CPU, cpuMutex *sync.Mutex) *Debugger {
	return &Debugger{
		CPU:         sim,
		CPUMutex:    cpuMutex,
		StepChannel: make(chan bool),
	}
}

func (d *Debugger) drawText(renderer *sdl.Renderer, font *ttf.Font, text string, x int, y int) error {
	textSurf, err := font.RenderUTF8Blended(text, sdl.Color{0, 0, 0, 255})
	if err != nil {
		return err
	}
	defer textSurf.Free()

	textTex, err := renderer.CreateTextureFromSurface(textSurf)
	if err != nil {
		return err
	}
	defer textTex.Destroy()

	_, _, w, h, err := textTex.Query()
	if err != nil {
		return err
	}

	renderer.Copy(textTex, &sdl.Rect{0, 0, w, h}, &sdl.Rect{int32(x), int32(y), w, h})

	return nil
}

func (d *Debugger) Loop(callback func()) {
	var runningMutex sync.Mutex

	ttf.Init()

	font12, err := ttf.OpenFont("/Users/student/Library/Fonts/FiraCode-Regular.ttf", 12)
	if err != nil {
		panic(err)
	}

	sdl.Main(func() {
		var window *sdl.Window
		var renderer *sdl.Renderer
		var err error

		sdl.Do(func() {
			window, err = sdl.CreateWindow("Debugger", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 600, 300, sdl.WINDOW_OPENGL)
		})
		if err != nil {
			log.Fatalf("Failed to create window and surface: %s\n", err)
		}
		defer func() {
			sdl.Do(func() {
				window.Destroy()
			})
		}()

		sdl.Do(func() {
			renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
		})
		if err != nil {
			log.Fatalf("Failed to create renderer: %s\n", err)
		}
		defer func() {
			sdl.Do(func() {
				renderer.Destroy()
			})
		}()

		sdl.Do(func() {
			renderer.Clear()
		})

		callback()

		running := true

		lastPC := uint16(0xFFFF)
		info := cpu.InstructionInfo{}
		formattedParams := ""
		dirty := true

		// disassemblyAddress := []uint16{}
		// disassemblyResults := []cpu.InstructionInfo{}

		for running {
			sdl.Do(func() {
				for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
					switch event.(type) {
					case *sdl.KeyboardEvent:
						e := event.(*sdl.KeyboardEvent)
						if e.State == sdl.RELEASED {
							if e.Keysym.Sym == sdl.K_SPACE {
								if d.SingleStep {
									dirty = true
									d.StepChannel <- true
								}
							}
						}
					case *sdl.QuitEvent:
						runningMutex.Lock()
						running = false
						runningMutex.Unlock()
					}
				}
			})

			sdl.Do(func() {
				if dirty && d.SingleStep {
					renderer.Clear()
					renderer.SetDrawColor(255, 255, 255, 255)
					renderer.FillRect(&sdl.Rect{0, 0, 600, 300})

					renderer.SetDrawColor(230, 230, 230, 255)
					renderer.FillRect(&sdl.Rect{0, 0, 600, 40})

					d.CPUMutex.Lock()

					d.drawText(renderer, font12, "A: "+strconv.FormatUint(uint64(d.CPU.Registers.A), 10), 0, 0)
					d.drawText(renderer, font12, "B: "+strconv.FormatUint(uint64(d.CPU.Registers.B), 10), 80, 0)
					d.drawText(renderer, font12, "C: "+strconv.FormatUint(uint64(d.CPU.Registers.C), 10), 160, 0)
					d.drawText(renderer, font12, "D: "+strconv.FormatUint(uint64(d.CPU.Registers.D), 10), 240, 0)
					d.drawText(renderer, font12, "E: "+strconv.FormatUint(uint64(d.CPU.Registers.E), 10), 320, 0)
					d.drawText(renderer, font12, "H: "+strconv.FormatUint(uint64(d.CPU.Registers.H), 10), 400, 0)
					d.drawText(renderer, font12, "L: "+strconv.FormatUint(uint64(d.CPU.Registers.L), 10), 480, 0)
					d.drawText(renderer, font12, "A': "+strconv.FormatUint(uint64(d.CPU.ShadowRegisters.A), 10), 0, 12)
					d.drawText(renderer, font12, "B': "+strconv.FormatUint(uint64(d.CPU.ShadowRegisters.B), 10), 80, 12)
					d.drawText(renderer, font12, "C': "+strconv.FormatUint(uint64(d.CPU.ShadowRegisters.C), 10), 160, 12)
					d.drawText(renderer, font12, "D': "+strconv.FormatUint(uint64(d.CPU.ShadowRegisters.D), 10), 240, 12)
					d.drawText(renderer, font12, "E': "+strconv.FormatUint(uint64(d.CPU.ShadowRegisters.E), 10), 320, 12)
					d.drawText(renderer, font12, "H': "+strconv.FormatUint(uint64(d.CPU.ShadowRegisters.H), 10), 400, 12)
					d.drawText(renderer, font12, "L': "+strconv.FormatUint(uint64(d.CPU.ShadowRegisters.L), 10), 480, 12)

					d.drawText(renderer, font12, "Flags: "+fmt.Sprintf("%08b", d.CPU.Registers.Flag), 0, 24)
					d.drawText(renderer, font12, "PC: 0x"+fmt.Sprintf("%04X", d.CPU.PC), 160, 24)
					d.drawText(renderer, font12, "SP: 0x"+fmt.Sprintf("%04X", d.CPU.Registers.SP), 240, 24)

					if lastPC != d.CPU.PC {
						info, formattedParams, _ = cpu.DisassembleInstructionAt(d.CPU, d.CPU.PC)
						lastPC = d.CPU.PC
					}

					d.drawText(renderer, font12, info.Mnemonic+" "+formattedParams, 0, 40)

					d.CPUMutex.Unlock()

					renderer.Present()

					dirty = false
				}
				if dirty && !d.SingleStep {
					renderer.Clear()
					renderer.SetDrawColor(255, 255, 255, 255)
					renderer.FillRect(&sdl.Rect{0, 0, 600, 300})

					d.drawText(renderer, font12, "Running at full speed, debugger disabled", 0, 0)

					renderer.Present()

					dirty = false
				}

				sdl.Delay(1000 / 60)
			})
		}
	})
}