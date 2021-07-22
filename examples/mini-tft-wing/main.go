package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/st7735"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"

	"github.com/bgould/go-seesaw"
)

var (
	i2c = machine.I2C0
)

func main() {

	time.Sleep(2 * time.Second)
	println("configuring I2C")
	i2c.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_100KHZ,
		SCL:       machine.SCL_PIN,
		SDA:       machine.SDA_PIN,
	})

	println("configuring seesaw device")
	minitft := seesaw.NewMiniTFTWing(i2c)
	if err := minitft.Configure(); err != nil {
		println("error", err)
		failMsg(err.Error())
	}
	if err := minitft.ResetTFT(true); err != nil {
		failMsg("error resetting: " + err.Error())
	}
	if err := minitft.SetBacklight(0x0); err != nil {
		println("error setting backlight:", err)
	}

	println("configuring display")
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
	})
	display := st7735.New(machine.SPI0, machine.D13, machine.D6, machine.D5, machine.D13)
	display.Configure(st7735.Config{
		Rotation: st7735.ROTATION_270,
	})

	black := color.RGBA{0, 0, 0, 255}
	display.FillScreen(black)

	white := color.RGBA{255, 255, 255, 255}
	width, height := display.Size()
	_, _ = width, height

	lastButtons := seesaw.MiniTFTWingButtons(seesaw.TFTWING_BUTTON_ALL)
	for {
		println("polling buttons")
		buttons, err := minitft.ReadButtons()
		println("buttons", buttons)
		if err != nil {
			println("error reading buttons", err)
		} else {
			if buttons != lastButtons {
				var s = "[ "
				if buttons.Left() {
					s += "L "
				} else {
					s += "  "
				}
				if buttons.Up() {
					s += "U "
				} else {
					s += "  "
				}
				if buttons.Down() {
					s += "D "
				} else {
					s += "  "
				}
				if buttons.Right() {
					s += "R "
				} else {
					s += "  "
				}
				if buttons.Select() {
					s += "S "
				} else {
					s += "  "
				}
				if buttons.B() {
					s += "B "
				} else {
					s += "  "
				}
				if buttons.A() {
					s += "A "
				} else {
					s += "  "
				}
				s += "]"
				fmt.Printf("buttons: %s\r\n", s)

				display.FillRectangle(0, 60, width, 35, black)
				tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 90, []byte(s), white)
				lastButtons = buttons
			}
		}
	}

}

func failMsg(msg string) {
	for {
		println("fail:", msg)
		time.Sleep(1 * time.Second)
	}
}
