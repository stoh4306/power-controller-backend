package main

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L/home/stoh/Codes/power-controller-backend/go -lpwctrlbe
// #include "./cpp/pwctrl-wrapper.h"
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"unsafe"

	"github.com/gin-gonic/gin"
)

type CmdResult struct {
	Cmd string `json:"cmd"`
	Res string `json:"result"`
}

type McuResponse struct {
	data      string `json:"data"`
	state     string `json:"state"`
	exception string `json:"exception"`
}

var pwCtrlBe unsafe.Pointer

// Create an instance of power-controller cpp backend
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
	pwCtrlBe = C.createPwctrlBackend()
	defer C.deletePwctrlBackend(unsafe.Pointer(pwCtrlBe))

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

	router := gin.Default()

	router.GET("/set/:id/:cmd", setPower)
	router.GET("/initialize", initialize)
	router.GET("/get/:id", getPower)

	router.Run(":8080")

	/*
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
		//*/

	//// Destroy the instance of power-controller cpp backend
	//C.deletePwctrlBackend(unsafe.Pointer(pwCtrlBe))
}

func setPower(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := C.CString(string(chars))
	defer C.free(unsafe.Pointer(mesg))

	paramId := c.Param("id")
	paramCmd := c.Param("cmd")
	log.Printf("cmd=%v, bmid=%v", paramCmd, paramId)

	tmpCmd := paramCmd + paramId
	cmdStr := C.CString(tmpCmd)

	result := int(C.set_command(pwCtrlBe, cmdStr, mesg, 64, 100))
	fmt.Printf("Mesg : %v", C.GoString(mesg))
	C.free(unsafe.Pointer(cmdStr))

	var tmpResponse CmdResult
	tmpResponse.Cmd = tmpCmd
	tmpResponse.Res = C.GoString(mesg)

	var response McuResponse

	if result == 0 {
		response.data = "true"
		response.state = "success"
		response.exception = ""
		c.IndentedJSON(http.StatusOK, response)
	} else {
		response.data = "false"
		response.state = "success"
		response.exception = ""
		c.IndentedJSON(http.StatusInternalServerError, response)
	}
}

func getPower(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := C.CString(string(chars))
	defer C.free(unsafe.Pointer(mesg))

	paramId := c.Param("id")
	log.Printf("cmd=power-check, bmid=%v", paramId)

	tmpCmd := "C" + paramId
	cmdStr := C.CString(tmpCmd)

	result := int(C.set_command(pwCtrlBe, cmdStr, mesg, 64, 100))
	fmt.Printf("Mesg : %v", C.GoString(mesg))
	C.free(unsafe.Pointer(cmdStr))

	var tmpresponse CmdResult
	tmpresponse.Cmd = tmpCmd
	tmpresponse.Res = C.GoString(mesg)

	var response McuResponse

	if result == 0 {
		response.data = "true"
		response.state = "success"
		response.exception = ""
		c.IndentedJSON(http.StatusOK, response)
	} else {
		response.data = "false"
		response.state = "success"
		response.exception = ""
		c.IndentedJSON(http.StatusInternalServerError, response)
	}
}

func initialize(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := C.CString(string(chars))
	defer C.free(unsafe.Pointer(mesg))

	result := int(C.initialize_connection(pwCtrlBe))

	var tmpresponse CmdResult
	tmpresponse.Cmd = "initialize"
	tmpresponse.Res = strconv.Itoa(result)

	var response McuResponse

	if result == 0 {
		response.data = "true"
		response.state = "success"
		response.exception = ""
		c.IndentedJSON(http.StatusOK, response)
	} else {
		response.data = "false"
		response.state = "success"
		response.exception = ""
		c.IndentedJSON(http.StatusInternalServerError, response)
	}
}

func readPort(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := C.CString(string(chars))
	defer C.free(unsafe.Pointer(mesg))

	result := int(C.readSerialPort(unsafe.Pointer(pwCtrlBe), mesg, 64))
	fmt.Printf("Mesg : %v", C.GoString(mesg))

	var response CmdResult
	response.Cmd = "read"
	response.Res = C.GoString(mesg)

	if result == 0 {
		c.IndentedJSON(http.StatusOK, response)
	} else {
		c.IndentedJSON(http.StatusInternalServerError, response)
	}
}

func writePort(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := C.CString(string(chars))
	defer C.free(unsafe.Pointer(mesg))

	tmpCmd := c.Param("cmd")
	cmdStr := C.CString(tmpCmd)
	result := int(C.set_command(unsafe.Pointer(pwCtrlBe), cmdStr, mesg, 64, 100))
	fmt.Printf("Mesg : %v", C.GoString(mesg))
	C.free(unsafe.Pointer(cmdStr))

	var response CmdResult
	response.Cmd = tmpCmd
	response.Res = C.GoString(mesg)

	if result == 0 {
		c.IndentedJSON(http.StatusOK, response)
	} else {
		c.IndentedJSON(http.StatusInternalServerError, response)
	}
}
