package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

var logger = logrus.New()

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

type HealthResponse struct {
	Status string `json:"status"`
}

type PwCtrl struct {
	portPrefix         string
	readTimeOut        int
	readMinByte        int
	portName           string
	connectInitialized bool
	serialPortFound    bool
	serialPort         serial.Port
	reIntializing      bool
}

var debugginMode_ bool

// PwCtrl constructor
//func NewPwCtrl(prefix string, timeout int, minbyte int) *PwCtrl {
//	return &PwCtrl{prefix, timeout, minbyte}
//}

var pwCtrl PwCtrl

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

	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	logger.Info(
		"Started with parameters : " + args[1] + ", " + args[2] + ", " + args[3])

	pwCtrl = PwCtrl{
		portPrefix:  "/dev/ttyACM",
		readTimeOut: 10,
		readMinByte: 0,

		portName:           "",
		connectInitialized: false,
		serialPortFound:    false,
		reIntializing:      false,
	}

	//pwCtrl.printValues()

	fmt.Println("====")

	err := readInputs(args)
	if err != nil {
		fmt.Println("Error in reading inputs : ", err.Error())
		return
	}

	//err = pwCtrl.findSerialPort()
	_, err = pwCtrl.intializeConnection()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Set debuggin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		debugginMode_ = false
		logger.Info("Running in Release mode")
	} else {
		debugginMode_ = true
		logger.Info("Running in Debugging mode")
	}

	router := gin.Default()

	router.GET("/set/:id/:cmd", setPower)
	router.GET("/initialize", initialize)
	router.GET("/get/:id", getPower)

	router.GET("/health", healthCheck)

	router.Run(":8080")

	logger.Info("Listening :8080")

	/*err = pwCtrl.write([]byte("0"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	buff := make([]byte, 100)
	n, err := pwCtrl.read(buff)
	if err != nil {
		return
	}
	fmt.Println(buff[:n])

	return*/
}

func healthCheck(c *gin.Context) {
	var healthResponse HealthResponse

	healthResponse.Status = "healthy"

	c.IndentedJSON(http.StatusOK, healthResponse)
}

func setPower(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := string(chars)

	paramId := c.Param("id")
	paramCmd := c.Param("cmd")
	tmpCmd := paramCmd + paramId
	logger.Infof("Sent command : %v", tmpCmd)

	err := pwCtrl.setCommand(tmpCmd, mesg, 100)
	logger.Infof("MCU response : %v", mesg)

	var tmpResponse CmdResult
	tmpResponse.Cmd = tmpCmd
	tmpResponse.Res = mesg

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

	if err == nil {
		response.State = "success"
		c.IndentedJSON(http.StatusOK, response)
	} else {
		var failResponse McuResponseFail
		failResponse.State = "fail"
		failResponse.Message = "Error"
		failResponse.ErrorType = "Unclassified"
		if mcuCode == 9 {
			c.IndentedJSON(http.StatusBadRequest, failResponse)
		} else {
			c.IndentedJSON(http.StatusInternalServerError, failResponse)
		}
	}
}

func getPower(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := string(chars)

	paramId := c.Param("id")
	tmpCmd := "C" + paramId
	logger.Infof("Sent command : %v", tmpCmd)

	err := pwCtrl.setCommand(tmpCmd, mesg, 100)
	logger.Infof("MCU response : %v", mesg)

	var tmpResponse CmdResult
	tmpResponse.Cmd = tmpCmd
	tmpResponse.Res = mesg

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

	if err == nil {
		response.State = "success"
		c.IndentedJSON(http.StatusOK, response)
	} else {
		var failResponse McuResponseFail
		failResponse.State = "fail"
		failResponse.Message = "Error"
		failResponse.ErrorType = "Unclassified"
		if mcuCode == 9 {
			c.IndentedJSON(http.StatusBadRequest, failResponse)
		} else {
			c.IndentedJSON(http.StatusInternalServerError, failResponse)
		}
	}
}

func initialize(c *gin.Context) {
	chars := make([]byte, 64)
	portName := string(chars)

	portName, err := pwCtrl.intializeConnection()
	if err != nil {
		logger.Info("Failed to initialize serial port")
	} else {
		logger.Info("Successfully initialized serial port : %v", portName)
	}

	if err != nil {
		var failResponse McuResponseFail
		failResponse.State = "fail"
		failResponse.Message = "Error"
		failResponse.ErrorType = "Unclassified"
		c.IndentedJSON(http.StatusInternalServerError, failResponse)
	} else {
		var response McuResponse
		response.Data = true
		response.State = "success"
		response.ElapsedSeconds = 0
		c.IndentedJSON(http.StatusOK, response)
	}
}

func (pwctl *PwCtrl) setCommand(cmdStr string, response string, sleepUTime int) error {
	err := pwctl.write([]byte(cmdStr))
	if err != nil {
		//TODO : re-initialzation code here
		if pwctl.reIntializing == false {
			go pwctl.reIntializeConnection()
		}
		return err
	} else {
		_, err = pwctl.read([]byte(response))
		if err == nil {
			return nil
		}

		if len(response) == 0 {
			logger.Info("[SERIAL-COM] ERROR, no response from MCU")
		} else if response[len(response)-1] != '\n' {
			logger.Info("[SERIAL-COM] Warning, incomplete response")
		} else if response[0] == '8' {
			logger.Info("[SERIAL-COM] ERROR, failed to power on/off")
		} else if response[0] == '9' {
			logger.Info("[SERIAL-COM] ERROR, unknown command or wrong rack-number")
		}

		return err
	}
}

func (pwctl *PwCtrl) reIntializeConnection() {
	// To prevent multiple executions of re-initializing
	if pwctl.reIntializing == false {
		pwctl.reIntializing = true

		for true {
			_, err := pwctl.intializeConnection()
			if err == nil {
				pwctl.reIntializing = false
				break
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func (pwctl *PwCtrl) read(buff []byte) (int, error) {
	n, err := pwctl.serialPort.Read(buff)
	if err != nil {
		return 0, err
	}
	fmt.Println("Data received : size = ", n)
	fmt.Println(string(buff[:n]))
	return n, nil
}

func (pwctl *PwCtrl) write(data []byte) error {
	n, err := pwctl.serialPort.Write(data)
	if err != nil {
		return err
	}

	fmt.Println("- Write data : size = ", n)
	return nil
}

func (pwctl *PwCtrl) intializeConnection() (string, error) {

	err := pwctl.findSerialPort()
	if err != nil {
		return "", err
	}

	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	pwctl.serialPort, err = serial.Open(pwctl.portName, mode)
	if err != nil {
		return pwctl.portName, err
	}

	pwctl.serialPort.SetReadTimeout(time.Duration(pwctl.readTimeOut) * time.Second)

	err = pwctl.serialPort.ResetInputBuffer()
	if err != nil {
		return pwctl.portName, err
	}

	err = pwctl.serialPort.ResetOutputBuffer()
	if err != nil {
		return pwctl.portName, err
	}

	fmt.Println("- Serial port initialized : ", pwctl.portName)

	return pwctl.portName, nil
}

func (p *PwCtrl) findSerialPort() error {
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

	return nil
}

func (p *PwCtrl) printValues() {
	fmt.Println("prefix:", p.portPrefix)
	fmt.Println("readTimeOut:", p.readTimeOut)
	fmt.Println("readMinByte:", p.readMinByte)
	fmt.Println("portName:", p.portName)
	fmt.Println("connectInitialized: ", p.connectInitialized)
	fmt.Println("portFound:", p.serialPortFound)
}

func readInputs(args []string) error {
	pwCtrl.portPrefix = "/dev/" + args[1]
	fmt.Println("portPrefix = ", pwCtrl.portPrefix)

	var err error
	pwCtrl.readTimeOut, err = strconv.Atoi(args[2])
	if err != nil {
		return err
	}
	fmt.Println("readTimeOut = ", pwCtrl.readTimeOut)

	pwCtrl.readMinByte, err = strconv.Atoi(args[3])
	if err != nil {
		return err
	}
	fmt.Println("readMinByte = ", pwCtrl.readMinByte)

	return nil
}
