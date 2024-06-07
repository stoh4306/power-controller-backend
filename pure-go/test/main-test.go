package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.bug.st/serial"
)

var portPrefix_ string
var portName_ string
var serialPort_ serial.Port
var readTimeOut_ int
var readMinByte_ int

func main() {
	fmt.Println("******************************")
	fmt.Println("Power-Controller-Backend")
	fmt.Println("******************************")

	args := os.Args

	if len(args) < 4 {
		fmt.Println("- usage : pwctl <arg1> <arg2> <arg3>")
		fmt.Println(" . arg1 : port name prefix (ex: ttyACM or ttyUSB)")
		fmt.Println(" . arg2 : max. reading time seconds")
		fmt.Println(" . arg3 : minimum bytes to read")
		fmt.Println(" . (example) pwctl ttyACM 10 0")
		return
	}

	// Read inputs
	err := readInputs(args)
	if err != nil {
		fmt.Println("ERROR : ", err)
		return
	}

	// Find the serial port with the input prefix
	err = findSerialPort()
	if err != nil {
		fmt.Println("ERROR : ", err)
		return
	}

	// Initialize port for serial communication
	serialPort_ = nil
	err = initializePort()
	if err != nil {
		fmt.Println("ERROR : ", err)
		return
	}

	for {
		var inCmd string
		fmt.Println("- Enter a command : (ex) C0000")
		fmt.Scanln(&inCmd)
		inCmd = inCmd + "\n"
		//fmt.Println(([]byte)(inCmd))

		if len(inCmd) > 4 && string(inCmd)[:4] == "exit" {
			break
		}

		// Write command to port
		serialPort_.ResetInputBuffer()

		_, err := serialPort_.Write([]byte(inCmd))
		if err != nil {
			fmt.Println("ERROR : ", err)
			return
		}

		// Read MCU reponse from port
		time.Sleep(200 * time.Millisecond)

		byteResponse := make([]byte, 64)
		n, err := serialPort_.Read(byteResponse)
		if err != nil {
			fmt.Println("ERROR : ", err)
			return
		}

		fmt.Println("- MCU Response : ")
		fmt.Println("  . in bytes  : #(bytes)=", n, byteResponse[:n])
		fmt.Println("  . in string : ", string(byteResponse[:n]))

		if n < 3 || n != 3 || byteResponse[1] != 13 || byteResponse[2] != 10 {
			fmt.Println("WARNING : Incomplete response!!!!")
		}
	}

	serialPort_.Close()
}

func initializePort() error {
	if serialPort_ != nil {
		serialPort_.Close()
	}

	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	var err error
	serialPort_, err = serial.Open(portName_, mode)
	if err != nil {
		return err
	}

	serialPort_.SetReadTimeout(time.Duration(readTimeOut_) * time.Second)

	err = serialPort_.ResetInputBuffer()
	if err != nil {
		return err
	}

	err = serialPort_.ResetOutputBuffer()
	if err != nil {
		return err
	}

	fmt.Println("- Serial port initialized : ", portName_)

	return nil
}

func findSerialPort() error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return err
	}

	portList := make([]string, 0)

	for _, port := range ports {
		if port[:len(portPrefix_)] == portPrefix_ {
			portList = append(portList, port)
		}
	}

	if len(portList) == 0 {
		return errors.New("no serial ports found")
	} else if len(portList) > 1 {
		fmt.Printf("Multiple serial port found :")
		for _, pn := range portList {
			fmt.Println("    - ", pn)
		}
		return errors.New("multiple serial ports found")
	}

	portName_ = portList[0]
	fmt.Printf("- Serial port found : %v\n", portName_)

	return nil
}

func printInputValues() {
	fmt.Println("*** Input parameters : ")
	fmt.Println(" - prefix:", portPrefix_)
	fmt.Println(" - readTimeOut:", readTimeOut_)
	fmt.Println(" - readMinByte:", readMinByte_)
	fmt.Println("***********************")
}

func readInputs(args []string) error {
	portPrefix_ = "/dev/" + args[1]

	var err error
	readTimeOut_, err = strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	readMinByte_, err = strconv.Atoi(args[3])
	if err != nil {
		return err
	}

	printInputValues()

	return nil
}
