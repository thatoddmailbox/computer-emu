package main

import (
	"bufio"
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/thatoddmailbox/minemu/bus"
	"github.com/thatoddmailbox/minemu/cpu"
	"github.com/thatoddmailbox/minemu/debugger"
	"github.com/thatoddmailbox/minemu/devices"
)

var st7565p *devices.ST7565P

func loadHexFile(path string, rom *devices.I2716) {
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
				rom.ROM[address+i] = uint8(dataByte)
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

func loadBinFile(path string, rom []byte) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.Read(rom)
}

func main() {
	log.Println("minemu")

	rand.Seed(time.Now().UTC().UnixNano())

	weirdMapping := flag.Bool("weird-mapping", false, "Enables the weird mapping, with two modern ROMs in ROM0 and ROM1, and a modern RAM chip in ROM3.")
	randomRam := flag.Bool("random-ram", false, "Fills the RAM with random data.")

	flag.Parse()

	bus := bus.EmulatorBus{}

	sim := cpu.CPU{}
	sim.Bus = bus
	cpuMutex := sync.Mutex{}

	dbg := debugger.NewDebugger(&sim, &cpuMutex, func() {
		log.Println("Execution resumed!")
		st7565p.PausedForBreakpoint = false
	})
	// dbg.SingleStep = true

	dbg.Loop(func() {
		if !*weirdMapping {
			// rom
			rom0 := devices.NewI2716(0x0000)
			rom1 := devices.NewI2716(0x1000)
			rom2 := devices.NewI2716(0x2000)
			rom3 := devices.NewI2716(0x3000)

			loadBinFile("bank0.bin", rom0.ROM[:])
			loadBinFile("bank1.bin", rom1.ROM[:])
			loadBinFile("bank2.bin", rom2.ROM[:])
			loadBinFile("bank3.bin", rom3.ROM[:])

			sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, rom0)
			sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, rom1)
			sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, rom2)
			sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, rom3)

			// ram
			ram := devices.NewKR537RU2()
			sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, ram)
		} else {
			rom0 := devices.NewSST39SF010A()
			rom1 := devices.NewAT28C256()

			loadBinFile("rom0.bin", rom0.ROM[:])
			loadBinFile("rom1.bin", rom1.ROM[:])

			sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, rom0)
			sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, rom1)

			ram := devices.NewAS6C62256()
			sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, ram)

			if *randomRam {
				for i := 0; i < 4*1024; i++ {
					ram.RAM[i] = uint8(rand.Uint32() % 256)
				}
			}
		}

		// peripherals
		pio := devices.NewI8255()
		st7565p = devices.NewST7565P(pio)
		sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, st7565p)
		sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, devices.NewI8251())
		sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, pio)

		// start the cpu
		go cpuRoutine(&sim, &cpuMutex, dbg)
	})
}

func cpuRoutine(sim *cpu.CPU, cpuMutex *sync.Mutex, dbg *debugger.Debugger) {
	defer (func() {
		err := recover()
		if err != nil {
			log.Println("PANIC")
			log.Printf("PC: 0x%x", sim.PC)

			info, disassembly, bytes := cpu.DisassembleInstructionAt(sim, sim.PC)

			log.Printf("%s %s (%d bytes)", info.Mnemonic, disassembly, bytes)
			log.Printf("Registers: %+v", sim.Registers)
			log.Println(err)
		}
	})()

	cycle := 0
	for {
		if dbg.SingleStep {
			<-dbg.StepChannel
		}

		cpuMutex.Lock()
		// info, disassembly, bytes := cpu.DisassembleInstructionAt(sim, sim.PC)

		// log.Printf("%s %s (%d bytes)", info.Mnemonic, disassembly, bytes)
		// log.Printf("%+v", sim.Registers)

		err := sim.Step(func() {
			log.Println("Breakpoint triggered!")
			st7565p.PausedForBreakpoint = true
			dbg.SingleStep = true
		})
		if err != nil {
			panic(err)
		}
		cpuMutex.Unlock()

		time.Sleep(time.Second / 10000000)

		if !dbg.SingleStep {
			cycle += 1
			if cycle > 1000 {
				cycle = 0
			}
		}
	}
}
