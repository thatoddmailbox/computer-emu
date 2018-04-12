package io

import (
	"fmt"
)

type UART struct {
	
}

func NewUART() *UART {
	return &UART{}
}

func (u *UART) IsMapped(address uint8) bool {
	if (address & (1 << 7) == 0) && (address & (1 << 6) == 0) {
		return true
	}
	return false
}

func (u *UART) ReadByte(address uint8) uint8 {
	maskedAddress := address & 1
	if maskedAddress == 0 {
		// data
		
	} else {
		// control/status
	}
	return 0xFF
}

func (u *UART) WriteByte(address uint8, data uint8) {
	maskedAddress := address & 1
	if maskedAddress == 0 {
		// data
		fmt.Print(string(data))
	} else {
		// control/status
	}
}