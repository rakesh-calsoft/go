// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package fsp provides access to the power supply unit
package fsp

import (
	"unsafe"

	"github.com/platinasystems/go/gpio"
	"github.com/platinasystems/go/i2c"
)

var (
	dummy       byte
	regsPointer = unsafe.Pointer(&dummy)
	regsAddr    = uintptr(unsafe.Pointer(&dummy))
)

type Psu struct {
	Bus        int
	Addr       int
	MuxAddr    int
	MuxValue   int
	GpioPwrok  string
	GpioPrsntL string
	GpioPwronL string
	GpioIntL   string
}

const ()

func getPsuRegs() *psuRegs { return (*psuRegs)(regsPointer) }
func getGenRegs() *genRegs { return (*genRegs)(regsPointer) }

// offset function has divide by two for 16-bit offset struct
func (r *reg8) offset() uint8   { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }
func (r *reg16) offset() uint8  { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }
func (r *reg16r) offset() uint8 { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }

func (h *Psu) i2cDo(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(h.Bus)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(h.Addr)
	if err != nil {
		return
	}

	err = bus.Do(rw, regOffset, size, data)
	return
}

func (h *Psu) i2cDoMux(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(h.Bus)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(h.MuxAddr)
	if err != nil {
		return
	}

	err = bus.Do(rw, regOffset, size, data)
	return
}

func (r *reg8) get(h *Psu) byte {
	var data i2c.SMBusData
	err := h.i2cDo(i2c.Read, r.offset(), i2c.ByteData, &data)
	if err != nil {
		panic(err)
	}
	return data[0]
}

func (r *reg8) setErr(h *Psu, v uint8) error {
	var data i2c.SMBusData
	data[0] = v
	return h.i2cDo(i2c.Write, r.offset(), i2c.ByteData, &data)
}

func (r *reg8) set(h *Psu, v uint8) {
	err := r.setErr(h, v)
	if err != nil {
		panic(err)
	}
}

func (r *reg8) setMux(h *Psu) error {
	var data i2c.SMBusData
	data[0] = byte(h.MuxValue)
	return h.i2cDoMux(i2c.Write, r.offset(), i2c.ByteData, &data)
}

func (r *reg16) get(h *Psu) (v uint16) {
	var data i2c.SMBusData
	err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
	return uint16(data[0])<<8 | uint16(data[1])
}

func (r *reg16r) get(h *Psu) (v uint16) {
	var data i2c.SMBusData
	err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
	return uint16(data[1])<<8 | uint16(data[0])
}

func (r *reg16) set(h *Psu, v uint16) {
	var data i2c.SMBusData
	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
}

func (r *reg16r) set(h *Psu, v uint16) {
	var data i2c.SMBusData
	data[1] = uint8(v >> 8)
	data[0] = uint8(v)
	err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
}

func (r *regi16) get(h *Psu) (v int16) { v = int16((*reg16)(r).get(h)); return }
func (r *regi16) set(h *Psu, v int16)  { (*reg16)(r).set(h, uint16(v)) }

func (h *Psu) Page() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Page.get(h)
	return (uint16(t))
}

func (h *Psu) PageWr(i uint16) {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	r.Page.set(h, uint8(i))
	return
}

func (h *Psu) StatusWord() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.StatusWord.get(h)
	return (uint16(t))
}

func (h *Psu) StatusVout() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.StatusVout.get(h)
	return (uint16(t))
}

func (h *Psu) StatusIout() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.StatusIout.get(h)
	return (uint16(t))
}

func (h *Psu) StatusInput() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.StatusInput.get(h)
	return (uint16(t))
}

func (h *Psu) StatusTemp() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.StatusTemp.get(h)
	return (uint16(t))
}

func (h *Psu) StatusFans() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.StatusFans.get(h)
	return (uint16(t))
}

func (h *Psu) Vin() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Vin.get(h)
	return (uint16(t))
}

func (h *Psu) Iin() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Iin.get(h)
	return (uint16(t))
}

func (h *Psu) Vout() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Vout.get(h)
	return (uint16(t))
}

func (h *Psu) Iout() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Iout.get(h)
	return (uint16(t))
}

func (h *Psu) Temp1() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Temp1.get(h)
	return (uint16(t))
}

func (h *Psu) Temp2() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Temp2.get(h)
	return (uint16(t))
}

func (h *Psu) FanSpeed() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.FanSpeed.get(h)
	return (uint16(t))
}

func (h *Psu) Pout() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Pout.get(h)
	return (uint16(t))
}

func (h *Psu) Pin() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.Pin.get(h)
	return (uint16(t))
}

func (h *Psu) PMBusRev() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.PMBusRev.get(h)
	return (uint16(t))
}

func (h *Psu) MfgId() uint16 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPsuRegs()
	t := r.MfgId.get(h)
	return (uint16(t))
}

func (h *Psu) PsuStatus() string {
	pin, found := gpio.Pins[h.GpioPrsntL]
	if !found {
		return "not_found"
	} else {
		t, err := pin.Value()
		if err != nil {
			return err.Error()
		} else if !t {
			return "not_installed."
		}
	}

	pin, found = gpio.Pins[h.GpioPwrok]
	if !found {
		return "undetermined"
	}
	t, err := pin.Value()
	if err != nil {
		return err.Error()
	}
	if !t {
		return "powered_off"
	}
	return "powered_on"
}

func (h *Psu) SetAdminState(s string) {
	pin, found := gpio.Pins[h.GpioPwronL]
	if found {
		switch s {
		case "disable":
			pin.SetValue(false)
		case "enable":
			pin.SetValue(true)
		}
	}
}

func (h *Psu) GetAdminState() string {
	pin, found := gpio.Pins[h.GpioPwronL]
	if !found {
		return "not found"
	}
	t, err := pin.Value()
	if err != nil {
		return err.Error()
	}
	if !t {
		return "disabled"
	}
	return "enabled"
}