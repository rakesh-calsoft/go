// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package i2cd

import (
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
)

const (
	Name    = "i2cd"
	Apropos = "i2c server daemon"
	Usage   = "i2cd"
)

func New() Command { return make(Command) }

type Command chan struct{}

func (Command) Apropos() lang.Alt { return apropos }

func (c Command) Close() error {
	close(c)
	return nil
}

func (Command) Kind() cmd.Kind { return cmd.Daemon }

func (c Command) Main(...string) error {
	var si syscall.Sysinfo_t
	err := syscall.Sysinfo(&si)
	if err != nil {
		return err
	}

	i2cReq := new(I2cReq)
	rpc.Register(i2cReq)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1233")
	if e != nil {
		log.Print("listen error:", e)
	}
	log.Print("listen OKAY")
	go http.Serve(l, nil)

	t := time.NewTicker(20 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-c:
			return nil
		case <-t.C:
		}
	}
	return nil
}

func (Command) String() string { return Name }
func (Command) Usage() string  { return Usage }

const MAXOPS = 30

type I struct {
	InUse     bool
	RW        i2c.RW
	RegOffset uint8
	BusSize   i2c.SMBusSize
	Data      [34]byte
	Bus       int
	Addr      int
	Delay     int
}
type R struct {
	D [34]byte
	E error
}

type I2cReq int

var b = [34]byte{0}
var i = I{false, i2c.RW(0), 0, 0, b, 0, 0, 0}
var j [MAXOPS]I
var r = R{b, nil}
var s [MAXOPS]R
var x int
var stopped byte = 0
var mutex = &sync.Mutex{}

func (t *I2cReq) ReadWrite(g *[MAXOPS]I, f *[MAXOPS]R) error {
	mutex.Lock()
	defer mutex.Unlock()

	var bus i2c.Bus
	var data i2c.SMBusData
	if g[0].Bus == 0x99 {
		stopped = byte(g[0].Addr)
		return nil
	}
	if g[0].Bus == 0x98 {
		f[0].D[0] = stopped
		return nil
	}
	for x := 0; x < MAXOPS; x++ {
		if g[x].InUse == true {
			err := bus.Open(g[x].Bus)
			if err != nil {
				log.Print("Error opening I2C bus")
				return err
			}
			defer bus.Close()

			err = bus.ForceSlaveAddress(g[x].Addr)
			if err != nil {
				log.Print("ERR2")
				log.Print("Error setting I2C slave address")
				return err
			}
			data[0] = g[x].Data[0]
			data[1] = g[x].Data[1]
			data[2] = g[x].Data[2]
			data[3] = g[x].Data[3]
			err = bus.Do(g[x].RW, g[x].RegOffset, g[x].BusSize, &data)
			if err != nil {
				log.Printf("Error doing I2C R/W: bus 0x%x addr 0x%x offset 0x%x data 0x%x RW %d BusSize %d delay %d", g[x].Bus, g[x].Addr, g[x].RegOffset, data[0], g[x].RW, g[x].BusSize, g[x].Delay)
				return err
			}
			f[x].D[0] = data[0]
			f[x].D[1] = data[1]
			if g[x].BusSize == i2c.I2CBlockData {
				for y := 2; y < 34; y++ {
					f[x].D[y] = data[y]
				}
			}
			bus.Close()
			if g[x].Delay > 0 {
				time.Sleep(time.Duration(g[x].Delay) * time.Millisecond)
			}
		}
	}
	return nil
}

func clearJS() {
	x = 0
	for k := 0; k < MAXOPS; k++ {
		j[k] = i
		s[k] = r
	}
}

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
