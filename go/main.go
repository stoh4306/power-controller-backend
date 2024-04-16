package main

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L/home/stoh/Codes/power-controller-backend/go -lpwctrlbe
// #include "./cpp/pwctrl-wrapper.h"
// #include <stdlib.h>
import "C"

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"unsafe"
)

func main() {
	args := os.Args

	if len(args) < 4 {
		fmt.Println("******************************")
		fmt.Println("Power-Controller-Backend")
		fmt.Println("- usage : pwctl <arg1> <arg2> <arg3>")
		fmt.Println(" . arg1 : port name prefix (ex: ttyACM or ttyUSB)")
		fmt.Println(" . arg2 : max. reading time in deciseconds (10decisec = 1sec)")
		fmt.Println(" . arg3 : minimum bytes to read")
		fmt.Println(" . (example) pwctl ttyACM 100 0")
		return
	}

	fmt.Println(args[1])
	fmt.Println(args[2])
	fmt.Println(args[3])

	portNamePrefix := C.CString(args[1])
	defer C.free(unsafe.Pointer(portNamePrefix))

	maxReadTime, _ := strconv.Atoi(args[2])
	minByte, _ := strconv.Atoi(args[3])

	// Create an instance of power-controller cpp backend
	pwCtrlBe := C.createPwctrlBackend()

	// Set port name prefix
	C.setPortNamePrefix(unsafe.Pointer(pwCtrlBe), portNamePrefix)

	// Set max reading time
	C.setMaxReadTime(unsafe.Pointer(pwCtrlBe), C.int(maxReadTime))

	// Set minimum bytes to read
	C.setMinimumBytes(unsafe.Pointer(pwCtrlBe), C.int(minByte))

	// Initialize connection
	result := int(C.initialize_connection(unsafe.Pointer(pwCtrlBe)))
	if result > 0 {
		fmt.Println("ERROR, failed to initialize connection")
		return
	}

	keyReader := bufio.NewReader(os.Stdin)
	chars := make([]byte, 255)
	mesg := C.CString(string(chars))
	defer C.free(unsafe.Pointer(mesg))

	for {
		fmt.Print("CMD :")
		cmdMesg, _ := keyReader.ReadString('\n')

		if cmdMesg[:1] == "r" {
			result = int(C.readSerialPort(unsafe.Pointer(pwCtrlBe), mesg, 256))
			fmt.Printf("Mesg : %v", C.GoString(mesg))
		} else if cmdMesg[:1] == "w" {
			tmpCmd := cmdMesg[1:]
			cmdStr := C.CString(tmpCmd)
			result = int(C.set_command(unsafe.Pointer(pwCtrlBe), cmdStr, mesg, 256, 100))
			fmt.Printf("Mesg : %v", C.GoString(mesg))
			C.free(unsafe.Pointer(cmdStr))
		} else if cmdMesg[:1] == "q" {
			//C.deletePwctrlBackend(unsafe.Pointer(pwCtrlBe))
			break
		} else {
			fmt.Printf("Unknown command : %v", cmdMesg[:1])
		}
	}

	// Destroy the instance of power-controller cpp backend
	C.deletePwctrlBackend(unsafe.Pointer(pwCtrlBe))
}
