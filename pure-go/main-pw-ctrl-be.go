package main

import (
	"fmt"
	"os"
	"strconv"
)

type PwCtrlBe struct {
	portPrefix  string
	readTimeOut int
	readMinByte int
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

	return
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
