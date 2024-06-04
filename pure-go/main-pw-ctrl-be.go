package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"go.bug.st/serial"
)

type PwCtrlBe struct {
	portPrefix  string
	readTimeOut int
	readMinByte int

	portName           string
	connectInitialized bool
}

// PwCtrlBe constructor
//func NewPwCtrlBe(prefix string, timeout int, minbyte int) *PwCtrlBe {
//	return &PwCtrlBe{prefix, timeout, minbyte}
//}

var pwCtrlBe PwCtrlBe

func main() {
	fmt.Println("******************************")
	fmt.Println("Power-Controller-Backend")
	fmt.Println("******************************")

	args := os.Args

	if len(args) < 4 {
		fmt.Println("- usage : pwctl <arg1> <arg2> <arg3>")
		fmt.Println(" . arg1 : port name prefix (ex: ttyACM or ttyUSB)")
		fmt.Println(" . arg2 : max. reading time in deciseconds (10decisec = 1sec)")
		fmt.Println(" . arg3 : minimum bytes to read")
		fmt.Println(" . (example) pwctl ttyACM 100 0")
		return
	}

	err := readInputs(args)
	if err != nil {
		fmt.Println("Error in reading inputs : ", err.Error())
		return
	}

	err = pwCtrlBe.findSerialPort()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	return
}

func (p *PwCtrlBe) intializeConnection() error {
	return nil
}

func (p *PwCtrlBe) findSerialPort() error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return err
	}

	portList := make([]string, 0)

	for _, port := range ports {
		if port[:len(p.portPrefix)] == p.portPrefix {
			portList = append(portList, port)
		}
	}

	if len(portList) != 1 {
		return errors.New("No or multiple ports found")
	}

	p.portName = portList[0]
	fmt.Printf("- Serial port found : %v\n", p.portName)

	/*mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.EvenParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}*/

	//serialPort, err := serial.Open(p.portName, mode)

	return nil
}

func readInputs(args []string) error {
	pwCtrlBe.portPrefix = "/dev/" + args[1]
	fmt.Println("portPrefix = ", pwCtrlBe.portPrefix)

	var err error
	pwCtrlBe.readTimeOut, err = strconv.Atoi(args[2])
	if err != nil {
		return err
	}
	fmt.Println("readTimeOut = ", pwCtrlBe.readTimeOut)

	pwCtrlBe.readMinByte, err = strconv.Atoi(args[3])
	if err != nil {
		return err
	}
	fmt.Println("readMinByte = ", pwCtrlBe.readMinByte)

	return nil
}
