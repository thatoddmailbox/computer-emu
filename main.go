package main

import (
    "bufio"
	// "io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/thatoddmailbox/minemu/bus"
	"github.com/thatoddmailbox/minemu/cpu"

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
				bus.ROM[address + i] = uint8(dataByte)
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
	log.Println("test")

	bus := bus.EmulatorBus{}

	loadBinFile("prg.bin", &bus)

	sim := cpu.CPU{}
	sim.Bus = bus

	for {
		instruction := sim.Bus.ReadMemoryByte(sim.PC)

		info := cpu.DisassemblyTable_Unprefixed[instruction]
		formattedParams := info.Parameters
		if info.DataBytes == 1 {
			data := uint64(sim.Bus.ReadMemoryByte(sim.PC + 1))
			formattedParams = strings.Replace(formattedParams, "%d8", "0x" + strconv.FormatUint(data, 16), -1)
		} else if info.DataBytes == 2 {
			data := uint64((uint16(sim.Bus.ReadMemoryByte(sim.PC + 2)) << 8) | uint16(sim.Bus.ReadMemoryByte(sim.PC + 1)))
			formattedParams = strings.Replace(formattedParams, "%d16", "0x" + strconv.FormatUint(data, 16), -1)
		}

		err := sim.Step()
		if err != nil {
			log.Printf("%s %s", info.Mnemonic, formattedParams)
			log.Printf("%+v", sim.Registers)
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
