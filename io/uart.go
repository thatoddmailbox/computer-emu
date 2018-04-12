package io

import (
	"bufio"
	"fmt"
	"os"
)

type UART struct {
	reader *bufio.Reader
}

func NewUART() *UART {
	return &UART{
		reader: bufio.NewReader(os.Stdin),
	}
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
		b, _ := u.reader.ReadByte()
		return b
	} else {
		// control/status
		flags := uint8(0)
		if u.reader.Buffered() > 0 {
			flags |= (1 << 1)
		}
		return flags
	}
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