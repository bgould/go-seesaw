package seesaw

import (
	"errors"
	"fmt"
	"time"
)

type Bus interface {
	Tx(addr uint16, w, r []byte) error
}

type FlowController interface {
	WaitUntilReady() bool
}

type FlowControllerFunc func() bool

func (fn FlowControllerFunc) WaitUntilReady() bool {
	return fn()
}

type Device struct {
	bus  Bus
	buf  []byte
	addr uint16
	flow FlowController
}

func NewDevice(bus Bus) *Device {
	return &Device{
		bus: bus,
		buf: make([]byte, 32),
	}
}

//var InvalidCodeError = errors.New("invalid HW ID code")

func (d *Device) Configure(address uint8, reset bool, fc FlowController) error {
	d.addr = uint16(address)
	d.flow = fc
	if d.flow == nil {
		d.flow = FlowControllerFunc(func() bool { return true })
	}
	if reset {
		if err := d.SoftwareReset(); err != nil {
			return fmt.Errorf("error during software reset: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	if c, err := d.ReadHardwareID(); err != nil {
		return fmt.Errorf("error reading HW ID code: %w", err)
	} else if c != SEESAW_HW_ID_CODE {
		return errors.New("invalid HW ID code") //InvalidCodeError
	} else {
		fmt.Printf("hardware code: %02x\n", c)
	}
	return nil
}

func (d *Device) ReadHardwareID() (byte, error) {
	return d.read8(SEESAW_STATUS_BASE, SEESAW_STATUS_HW_ID)
}

func (d *Device) SoftwareReset() error {
	return d.write8(SEESAW_STATUS_BASE, SEESAW_STATUS_SWRST, 0xFF)
}

func (d *Device) read8(hi byte, lo byte) (byte, error) {
	if err := d.read(hi, lo, d.buf[:1]); err != nil {
		return 0, err
	}
	return d.buf[0], nil
}

func (d *Device) read(hi byte, lo byte, buf []byte) error {
	d.buf[0] = hi
	d.buf[1] = lo
	println("read 1")
	if err := d.bus.Tx(d.addr, d.buf[0:2], nil); err != nil {
		println("error")
		return err
	}
	println("read 2")
	d.waitUntilReady()
	println("read 3")
	if err := d.bus.Tx(d.addr, nil, buf); err != nil {
		return err
	}
	return nil
}

func (d *Device) write8(hi byte, lo byte, val byte) error {
	d.buf[0] = hi
	d.buf[1] = lo
	d.buf[2] = val
	if err := d.bus.Tx(d.addr, d.buf[0:3], nil); err != nil {
		return err
	}
	d.waitUntilReady()
	return nil
}

func (d *Device) waitUntilReady() {
	if d.flow != nil {
		for !d.flow.WaitUntilReady() {
		}
	}
}

func (d *Device) write(hi byte, lo byte, buf []byte) error {
	d.buf[0] = hi
	d.buf[1] = lo
	buflen := len(buf)
	if copy(d.buf[2:], buf) != buflen {
		return errors.New("seesaw: buf too large")
	}
	if err := d.bus.Tx(d.addr, d.buf[0:buflen+2], nil); err != nil {
		return err
	}
	d.waitUntilReady()
	return nil
}

type PinMode uint8

const (
	PinInput       PinMode = 0x0
	PinOutput      PinMode = 0x1
	PinInputPullup PinMode = 0x2
)

func (d *Device) pinMode(pin uint8, mode PinMode) error {
	if pin >= 32 {
		return d.pinModeBulk(0, 1<<(pin-32), mode)
	} else {
		return d.pinModeBulk(1<<pin, 0, mode)
	}
}

func (d *Device) pinModeBulk(pinsA uint32, pinsB uint32, mode PinMode) error {
	cmd := []byte{
		byte(pinsA >> 24), byte(pinsA >> 16), byte(pinsA >> 8), byte(pinsA),
		byte(pinsB >> 24), byte(pinsB >> 16), byte(pinsB >> 8), byte(pinsB),
	}
	switch mode {
	case PinInput:
		return d.write(SEESAW_GPIO_BASE, SEESAW_GPIO_DIRCLR_BULK, cmd)
	case PinOutput:
		return d.write(SEESAW_GPIO_BASE, SEESAW_GPIO_DIRSET_BULK, cmd)
	case PinInputPullup:
		var err error
		err = d.write(SEESAW_GPIO_BASE, SEESAW_GPIO_DIRCLR_BULK, cmd)
		if err == nil {
			err = d.write(SEESAW_GPIO_BASE, SEESAW_GPIO_PULLENSET, cmd)
		}
		if err == nil {
			err = d.write(SEESAW_GPIO_BASE, SEESAW_GPIO_BULK_SET, cmd)
		}
		return err
	}
	return errors.New("unknown pinMode")
}

func (d *Device) digitalWrite(pin uint8, state bool) error {
	if pin >= 32 {
		return d.digitalWriteBulk(0, 1<<(pin-32), state)
	} else {
		return d.digitalWriteBulk(1<<pin, 0, state)
	}
}

func (d *Device) digitalWriteBulk(pinsA uint32, pinsB uint32, state bool) error {
	cmd := []byte{
		byte(pinsA >> 24), byte(pinsA >> 16), byte(pinsA >> 8), byte(pinsA),
		byte(pinsB >> 24), byte(pinsB >> 16), byte(pinsB >> 8), byte(pinsB),
	}
	if state {
		return d.write(SEESAW_GPIO_BASE, SEESAW_GPIO_BULK_SET, cmd)
	} else {
		return d.write(SEESAW_GPIO_BASE, SEESAW_GPIO_BULK_CLR, cmd)
	}
}

func (d *Device) digitalReadBulk(pins uint32) (uint32, error) {
	println("digital read bulk")
	b := d.buf[:4]
	if err := d.read(SEESAW_GPIO_BASE, SEESAW_GPIO_BULK, b); err != nil {
		return 0, err
	}
	println("after read")
	ret := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return ret & pins, nil
}
