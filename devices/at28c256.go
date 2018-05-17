package devices

import (
	"errors"
)

// in the rom0 slot

type AT28C256 struct {
	ROM      [32 * 1024]byte
	ReadOnly bool
}

func NewAT28C256() *AT28C256 {
	return &AT28C256{
		ReadOnly: true,
	}
}

func (r *AT28C256) IsMapped(address uint16) bool {
	if (address & 0xF000) == 0x1000 {
		return true
	} else {
		return false
	}
}

func (r *AT28C256) ReadByte(address uint16) uint8 {
	accessAddress := address & 0x0FFF

	// a13 and a11 held high, computer a11 copied to a12
	// a14 held low
	oldA11 := accessAddress & (1 << 11)
	accessAddress |= (1 << 13)
	accessAddress |= (1 << 11)
	if oldA11 != 0 {
		accessAddress |= (1 << 12)
	}

	return r.ROM[accessAddress]
}

func (r *AT28C256) WriteByte(address uint16, data uint8) {
	if r.ReadOnly {
		panic(errors.New("at28c256: write to ROM"))
	}

	accessAddress := address & 0x0FFF

	// a13 and a11 held high, computer a11 copied to a12
	// a14 held low
	oldA11 := accessAddress & (1 << 11)
	accessAddress |= (1 << 13)
	accessAddress |= (1 << 11)
	if oldA11 != 0 {
		accessAddress |= (1 << 12)
	}

	r.ROM[accessAddress] = data
}
