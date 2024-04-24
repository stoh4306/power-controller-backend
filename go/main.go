package main

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L. -lpwctrlbe
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
	"github.com/sirupsen/logrus"
)

type CmdResult struct {
	Cmd string `json:"cmd"`
	Res string `json:"result"`
}

type McuResponse struct {
	Data           bool   `json:"data"`
	State          string `json:"state"`
	ElapsedSeconds int    `json:"elapsedSeconds"`
}

type McuResponseFail struct {
	State     string `json:"state"`
	Message   string `json:"message"`
	ErrorType string `json:"errorType"`
}

type McuResponseAllInOne struct {
	Data             string `json:"data"`
	State            string `json:"state"`
	Message          string `json:"message"`
	ElapsedSeconds   string `json:"elapsedSeconds"`
	ErrorType        string `json:"errorType"`
	ErrorDescription string `json:"errorDescription"`
	Debug            string `json:"debug"`
}

var pwCtrlBe unsafe.Pointer
var healthStatus int
var logger = logrus.New()

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

	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	logger.Info("Started with parameters : " + args[1] + ", " + args[2] + " " + args[3])

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
	healthStatus = result
	if result > 0 {
		logger.Error("Failed to initialize serial port: code=" + strconv.Itoa(result))
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
	tmpCmd := paramCmd + paramId
	log.Printf("cmd=%v", tmpCmd)

	cmdStr := C.CString(tmpCmd)
	defer C.free(unsafe.Pointer(cmdStr))

	result := int(C.set_command(pwCtrlBe, cmdStr, mesg, 64, 100))
	//fmt.Printf("Mesg : %v", C.GoString(mesg))

	var tmpResponse CmdResult
	tmpResponse.Cmd = tmpCmd
	tmpResponse.Res = C.GoString(mesg)

	mcuCode := 0
	if len(tmpResponse.Res) > 0 {
		mcuCode, _ = strconv.Atoi(tmpResponse.Res[:1])
	}

	var response McuResponse
	response.ElapsedSeconds = 0

	if mcuCode == 0 || mcuCode == 2 {
		response.Data = false
	} else if mcuCode == 1 || mcuCode == 3 {
		response.Data = true
	} else {
		response.Data = false
	}

	if result == 0 {
		response.State = "success"
		c.IndentedJSON(http.StatusOK, response)
	} else {
		response.State = "fail"
		var failResponse McuResponseFail
		failResponse.State = "fail"
		failResponse.Message = "Error"
		failResponse.ErrorType = "Unclassified"
		c.IndentedJSON(http.StatusInternalServerError, failResponse)
	}
}

func getPower(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := C.CString(string(chars))
	defer C.free(unsafe.Pointer(mesg))

	paramId := c.Param("id")
	tmpCmd := "C" + paramId
	cmdStr := C.CString(tmpCmd)
	defer C.free(unsafe.Pointer(cmdStr))

	log.Printf("cmd=%v", tmpCmd)

	result := int(C.set_command(pwCtrlBe, cmdStr, mesg, 64, 100))
	//fmt.Printf("Mesg : %v", C.GoString(mesg))

	var tmpresponse CmdResult
	tmpresponse.Cmd = tmpCmd
	tmpresponse.Res = C.GoString(mesg)

	mcuCode := 0
	if len(tmpresponse.Res) > 0 {
		mcuCode, _ = strconv.Atoi(tmpresponse.Res[:1])
	}

	var response McuResponse
	response.ElapsedSeconds = 0

	if mcuCode == 0 || mcuCode == 2 {
		response.Data = false
	} else if mcuCode == 1 || mcuCode == 3 {
		response.Data = true
	} else {
		response.Data = false
	}

	if result == 0 {
		response.State = "success"
		c.IndentedJSON(http.StatusOK, response)
	} else {
		var failResponse McuResponseFail
		failResponse.State = "fail"
		failResponse.Message = "Error"
		failResponse.ErrorType = "Unclassified"
		c.IndentedJSON(http.StatusInternalServerError, failResponse)
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
		response.Data = true
		response.State = "success"
		response.ElapsedSeconds = 0
		c.IndentedJSON(http.StatusOK, response)
	} else {
		//response.Data = false
		//response.State = "fail"
		//response.ElapsedSeconds = 0
		var failResponse McuResponseFail
		failResponse.State = "fail"
		failResponse.Message = "Error"
		failResponse.ErrorType = "Unclassified"
		c.IndentedJSON(http.StatusInternalServerError, failResponse)
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
	defer C.free(unsafe.Pointer(cmdStr))

	result := int(C.set_command(unsafe.Pointer(pwCtrlBe), cmdStr, mesg, 64, 100))
	fmt.Printf("Mesg : %v", C.GoString(mesg))

	var response CmdResult
	response.Cmd = tmpCmd
	response.Res = C.GoString(mesg)

	if result == 0 {
		c.IndentedJSON(http.StatusOK, response)
	} else {
		c.IndentedJSON(http.StatusInternalServerError, response)
	}
}
