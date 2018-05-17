package devices

import (
	"bufio"
	"fmt"
	"os"
)

type I8251 struct {
	reader *bufio.Reader
}

func NewI8251() *I8251 {
	return &I8251{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (u *I8251) IsMapped(address uint16) bool {
	if (address&(1<<15) == 0) && (address&(1<<14) != 0) && (address&(1<<13) == 0) && (address&(1<<12) == 0) {
		return true
	}
	return false
}

func (u *I8251) ReadByte(address uint16) uint8 {
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
		flags |= (1 << 2) // txempty
		return flags
	}
}

func (u *I8251) WriteByte(address uint16, data uint8) {
	maskedAddress := address & 1
	if maskedAddress == 0 {
		// data
		fmt.Print(string(data))
	} else {
		// control/status
	}
}
