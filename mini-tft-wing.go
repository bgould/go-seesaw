package seesaw

import (
	"fmt"
	"time"
)

const TFTWING_ADDR = 0x5E
const TFTWING_RESET_PIN = 8

const TFTWING_BACKLIGHT_ON = 0       // inverted output!
const TFTWING_BACKLIGHT_OFF = 0xFFFF // inverted output!

const TFTWING_BUTTON_UP_PIN = 2
const TFTWING_BUTTON_UP = 1 << TFTWING_BUTTON_UP_PIN

const TFTWING_BUTTON_DOWN_PIN = 4
const TFTWING_BUTTON_DOWN = 1 << TFTWING_BUTTON_DOWN_PIN

const TFTWING_BUTTON_LEFT_PIN = 3
const TFTWING_BUTTON_LEFT = 1 << TFTWING_BUTTON_LEFT_PIN

const TFTWING_BUTTON_RIGHT_PIN = 7
const TFTWING_BUTTON_RIGHT = 1 << TFTWING_BUTTON_RIGHT_PIN

const TFTWING_BUTTON_SELECT_PIN = 11
const TFTWING_BUTTON_SELECT = 1 << TFTWING_BUTTON_SELECT_PIN

const TFTWING_BUTTON_A_PIN = 10
const TFTWING_BUTTON_A = 1 << TFTWING_BUTTON_A_PIN

const TFTWING_BUTTON_B_PIN = 9
const TFTWING_BUTTON_B = 1 << TFTWING_BUTTON_B_PIN

const TFTWING_BUTTON_ALL = TFTWING_BUTTON_UP | TFTWING_BUTTON_DOWN |
	TFTWING_BUTTON_LEFT | TFTWING_BUTTON_RIGHT | TFTWING_BUTTON_SELECT |
	TFTWING_BUTTON_A | TFTWING_BUTTON_B

type MiniTFTWing struct {
	dev *Device
}

func NewMiniTFTWing(bus Bus) *MiniTFTWing {
	return &MiniTFTWing{dev: NewDevice(bus)}
}

func (w *MiniTFTWing) Configure() error {
	if err := w.dev.Configure(TFTWING_ADDR, true, wait150us); err != nil {
		return err
	}
	if err := w.dev.pinMode(TFTWING_RESET_PIN, PinOutput); err != nil {
		return err
	}
	if err := w.dev.pinModeBulk(TFTWING_BUTTON_ALL, 0, PinInputPullup); err != nil {
		return err
	}
	return nil
}

func (w *MiniTFTWing) SetBacklight(value uint16) error {
	cmd := []byte{0x00, byte(value >> 8), byte(value)}
	return w.dev.write(SEESAW_TIMER_BASE, SEESAW_TIMER_PWM, cmd)
}

func (w *MiniTFTWing) SetBacklightFreq(value uint16) error {
	cmd := []byte{0x00, byte(value >> 8), byte(value)}
	return w.dev.write(SEESAW_TIMER_BASE, SEESAW_TIMER_FREQ, cmd)
}

func (w *MiniTFTWing) ResetTFT(reset bool) error {
	return w.dev.digitalWrite(TFTWING_RESET_PIN, reset)
}

func (w *MiniTFTWing) ReadButtons() (MiniTFTWingButtons, error) {
	buttons, err := w.dev.digitalReadBulk(TFTWING_BUTTON_ALL)
	return MiniTFTWingButtons(buttons), err
}

type MiniTFTWingButtons uint32

func (b MiniTFTWingButtons) Up() bool {
	return b&TFTWING_BUTTON_UP == 0
}

func (b MiniTFTWingButtons) Down() bool {
	return b&TFTWING_BUTTON_DOWN == 0
}

func (b MiniTFTWingButtons) Left() bool {
	return b&TFTWING_BUTTON_LEFT == 0
}

func (b MiniTFTWingButtons) Right() bool {
	return b&TFTWING_BUTTON_RIGHT == 0
}

func (b MiniTFTWingButtons) Select() bool {
	return b&TFTWING_BUTTON_SELECT == 0
}

func (b MiniTFTWingButtons) A() bool {
	return b&TFTWING_BUTTON_A == 0
}

func (b MiniTFTWingButtons) B() bool {
	return b&TFTWING_BUTTON_B == 0
}

func (b MiniTFTWingButtons) String() string {
	return fmt.Sprintf("[U:%t D:%t L:%t R:%t S:%t A:%t B:%t]",
		b.Up(), b.Down(), b.Left(), b.Right(), b.Select(), b.A(), b.B(),
	)
}

var wait150us = FlowControllerFunc(func() bool {
	for t := time.Now(); time.Since(t) < 150*time.Microsecond; {
	}
	return true
})
