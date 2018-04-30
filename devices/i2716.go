package devices

import (
	"errors"
)

type I2716 struct {
	baseAddress uint16
	ROM         [2 * 1024]byte
	ReadOnly    bool
}

func NewI2716(baseAddress uint16) *I2716 {
	return &I2716{
		baseAddress: baseAddress,
		ReadOnly:    false,
	}
}

func (r *I2716) IsMapped(address uint16) bool {
	if (address & 0xF000) == r.baseAddress {
		return true
	} else {
		return false
	}
}

func (r *I2716) ReadByte(address uint16) uint8 {
	accessAddress := address & 0x07FF
	return r.ROM[accessAddress]
}

func (r *I2716) WriteByte(address uint16, data uint8) {
	if r.ReadOnly {
		panic(errors.New("i2716: write to read-only ROM"))
	}
	accessAddress := address & 0x07FF
	r.ROM[accessAddress] = data
}
