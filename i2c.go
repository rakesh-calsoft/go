// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package i2c provides cli command to access i2c devices.
package i2c

import (
	"fmt"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/i2c"
	"github.com/platinasystems/oops"
	"time"
)

type I2c struct{ oops.Id }

var Commands = goes.Commands{
	&I2c{"i2c"},
}

func (*I2c) Usage() string {
	return "i2c BUS.ADDR[.REG] [VALUE] [WRITE-DELAY-IN-SEC]"
}

func (p *I2c) Main(ctx *goes.Context, args ...string) {
	var (
		bus     i2c.Bus
		sd      i2c.SMBusData
		b, a, d uint8
		cs      [2]uint8
	)

	if n := len(args); n == 0 {
		p.Panic("BUS.ADDR.REG: missing")
	} else if n > 3 {
		p.Panic(args[3:], ": unexpected")
	}

	dValid := len(args) > 1

	nc := 2
	_, err := fmt.Sscanf(args[0], "%x.%x.%x-%x", &b, &a, &cs[0], &cs[1])
	if err != nil {
		nc = 1
		_, err = fmt.Sscanf(args[0], "%x.%x.%x", &b, &a, &cs[0])
		if err != nil {
			nc = 0
			_, err = fmt.Sscanf(args[0], "%x.%x", &b, &a)
		}
	}
	if err != nil {
		p.Panic(args[0], ": invalid BUS.ADDR[.REG]: ", err)
	}

	if dValid {
		_, err = fmt.Sscanf(args[1], "%x", &d)
		if err != nil {
			p.Panic(args[1], ": invalid value: ", err)
		}
	}

	writeDelay := float64(0)
	if len(args) > 2 {
		s := args[2]
		_, err := fmt.Sscanf(s, "%f", &writeDelay)
		if err != nil {
			p.Panic(s, ": invalid delay: ", err)
		}
	}

	err = bus.Open(int(b))
	if err != nil {
		p.Panic(err)
	}
	defer bus.Close()
	err = bus.ForceSlaveAddress(int(a))
	if err != nil {
		p.Panic(err)
	}
	op := i2c.ByteData
	if nc == 0 {
		op = i2c.Byte
	}
	rw := i2c.Read
	if dValid {
		rw = i2c.Write
		sd[0] = d
	}
	if nc < 2 {
		cs[1] = cs[0]
	}
	c := cs[0]
	for {
		err = bus.Do(rw, c, op, &sd)
		if err != nil {
			p.Panic(err)
		}
		ctx.Printf("%x.%02x.%02x = %02x\n", b, a, c, sd[0])
		if c == cs[1] {
			break
		}
		c++
		if rw == i2c.Write && writeDelay > 0 {
			dt := time.Duration(1e9 * writeDelay)
			time.Sleep(dt)
		}
	}
}
