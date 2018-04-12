package main

import (
	"bufio"
	"log"
	"os"
	"strconv"

	"github.com/thatoddmailbox/minemu/bus"
	"github.com/thatoddmailbox/minemu/cpu"
	"github.com/thatoddmailbox/minemu/io"
	// "github.com/veandco/go-sdl2/sdl"
)

func loadHexFile(path string, bus *bus.EmulatorBus) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		byteCount, _ := strconv.ParseUint(scanner.Text()[1:3], 16, 8)
		address, _ := strconv.ParseUint(scanner.Text()[3:7], 16, 16)
		recordType, _ := strconv.ParseUint(scanner.Text()[7:9], 16, 8)
		if recordType == 0 {
			for i := uint64(0); i < byteCount; i += 1 {
				dataByte, _ := strconv.ParseUint(scanner.Text()[9+(i*2):11+(i*2)], 16, 8)
				bus.ROM[address+i] = uint8(dataByte)
			}
		} else if recordType == 1 {
			break
		} else {
			panic("bad hex file")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func loadBinFile(path string, bus *bus.EmulatorBus) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.Read(bus.ROM[:])
}

func main() {
	log.Println("minemu")

	bus := bus.EmulatorBus{}

	bus.DataDevices = append(bus.DataDevices, io.NewUART())

	loadBinFile("prg.bin", &bus)

	sim := cpu.CPU{}
	sim.Bus = bus

	defer (func() {
		err := recover()
		if err != nil {
			info, disassembly, bytes := cpu.DisassembleInstructionAt(sim, sim.PC)

			log.Println("PANIC")
			log.Printf("%s %s (%d bytes)", info.Mnemonic, disassembly, bytes)
			log.Printf("PC: 0x%x", sim.PC)
			log.Printf("Registers: %+v", sim.Registers)
			log.Println(err)
		}
	})()

	for {
		// info, disassembly, bytes := cpu.DisassembleInstructionAt(sim, sim.PC)

		// log.Printf("%s %s (%d bytes)", info.Mnemonic, disassembly, bytes)
		// log.Printf("%+v", sim.Registers)

		err := sim.Step()
		if err != nil {
			panic(err)
		}
	}
	log.Printf("%+v", sim.Registers)
}

// func sdlTest() {
// 	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
// 		panic(err)
// 	}
// 	defer sdl.Quit()

// 	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
// 		800, 600, sdl.WINDOW_SHOWN)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer window.Destroy()

// 	surface, err := window.GetSurface()
// 	if err != nil {
// 		panic(err)
// 	}
// 	surface.FillRect(nil, 0)

// 	rect := sdl.Rect{0, 0, 200, 200}
// 	surface.FillRect(&rect, 0xffff0000)
// 	window.UpdateSurface()

// 	running := true
// 	for running {
// 		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
// 			switch event.(type) {
// 			case *sdl.QuitEvent:
// 				println("Quit")
// 				running = false
// 				break
// 			}
// 		}
// 	}
// }
