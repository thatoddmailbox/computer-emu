package devices

import (
	"errors"
)

// in the rom0 slot

type SST39SF010A struct {
	ROM      [128 * 1024]byte
	ReadOnly bool
}

func NewSST39SF010A() *SST39SF010A {
	return &SST39SF010A{
		ReadOnly: true,
	}
}

func (r *SST39SF010A) IsMapped(address uint16) bool {
	if (address & 0xF000) == 0x0000 {
		return true
	} else {
		return false
	}
}

func (r *SST39SF010A) ReadByte(address uint16) uint8 {
	accessAddress := address & 0x0FFF

	// a13 and a11 held high, computer a11 copied to a12
	// a16, a15, a14 all held low
	oldA11 := accessAddress & (1 << 11)
	accessAddress |= (1 << 13)
	accessAddress |= (1 << 11)
	if oldA11 != 0 {
		accessAddress |= (1 << 12)
	}

	return r.ROM[accessAddress]
}

func (r *SST39SF010A) WriteByte(address uint16, data uint8) {
	if r.ReadOnly {
		panic(errors.New("sst39sf010a: write to ROM"))
	}

	accessAddress := address & 0x0FFF

	// a13 and a11 held high, computer a11 copied to a12
	// a16, a15, a14 all held low
	oldA11 := accessAddress & (1 << 11)
	accessAddress |= (1 << 13)
	accessAddress |= (1 << 11)
	if oldA11 != 0 {
		accessAddress |= (1 << 12)
	}

	r.ROM[accessAddress] = data
}
