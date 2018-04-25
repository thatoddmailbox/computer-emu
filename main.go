package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/thatoddmailbox/minemu/bus"
	"github.com/thatoddmailbox/minemu/cpu"
	"github.com/thatoddmailbox/minemu/debugger"
	"github.com/thatoddmailbox/minemu/devices"
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

	loadBinFile("bank0.bin", &bus)

	sim := cpu.CPU{}
	sim.Bus = bus
	cpuMutex := sync.Mutex{}

	dbg := debugger.NewDebugger(&sim, &cpuMutex)
	// dbg.SingleStep = true

	dbg.Loop(func() {
		pio := devices.NewI8255()
		sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, devices.NewST7565P(pio))
		sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, devices.NewI8251())
		sim.Bus.MemoryDevices = append(sim.Bus.MemoryDevices, pio)
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
			dbg.SingleStep = true
		})
		if err != nil {
			panic(err)
		}
		cpuMutex.Unlock()

		if !dbg.SingleStep {
			cycle += 1
			if cycle > 1000 {
				cycle = 0
			}
		}
	}
}
