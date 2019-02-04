package debugger

import (
	"fmt"
	"log"
	"os/user"
	"strconv"
	"sync"

	"github.com/thatoddmailbox/computer-emu/cpu"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type DebuggerState struct {
	CPU cpu.CPU
}

type Debugger struct {
	StepChannel      chan bool
	CPU              *cpu.CPU
	CPUMutex         *sync.Mutex
	SingleStep       bool
	breakpointResume func()
}

func NewDebugger(sim *cpu.CPU, cpuMutex *sync.Mutex, breakpointResume func()) *Debugger {
	return &Debugger{
		CPU:              sim,
		CPUMutex:         cpuMutex,
		StepChannel:      make(chan bool),
		breakpointResume: breakpointResume,
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

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	font12, err := ttf.OpenFont(user.HomeDir+"/Library/Fonts/FiraCode-Regular.ttf", 12)
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
		lastSingleStep := d.SingleStep

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
							} else if e.Keysym.Sym == sdl.K_r {
								if d.SingleStep {
									dirty = true
									d.SingleStep = false
									d.breakpointResume()
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

					// d.drawText(renderer, font12, fmt.Sprintf("dropping: 0x%x, fall index: %d, random: %d", d.CPU.Bus.ReadMemoryByte(0xF004), d.CPU.Bus.ReadMemoryByte(0xF007), d.CPU.Bus.ReadMemoryByte(0xF002)), 0, 64)
					// d.drawText(renderer, font12, "tetris board:", 0, 76)
					// for i := 0; i < 14; i++ {
					// 	d.drawText(renderer, font12, fmt.Sprintf("%08b", d.CPU.Bus.ReadMemoryByte(uint16(0xF02A+i))), 0, 88+(i*12))
					// }
					// d.drawText(renderer, font12, "tetris fall zone:", 200, 76)
					// for i := 0; i < (15 + 4 + 4); i++ {
					// 	x := 200
					// 	y := 88 + (i * 12)
					// 	prefix := ""
					// 	if i > 14 && i < 14+5 {
					// 		prefix = "* "
					// 	}
					// 	if i > 14 {
					// 		x = 300
					// 		y -= (13 * 14)
					// 	}
					// 	d.drawText(renderer, font12, fmt.Sprintf("%s%08b", prefix, d.CPU.Bus.ReadMemoryByte(uint16(0xF008+i))), x, y)
					// }

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

				if d.SingleStep != lastSingleStep {
					dirty = true
					lastSingleStep = d.SingleStep
				}

				sdl.Delay(1000 / 60)
			})
		}
	})
}
