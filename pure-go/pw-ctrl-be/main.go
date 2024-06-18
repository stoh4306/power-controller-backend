package main

import (
	"errors"
	"fmt"
	"grida/pwctrlbe/docs"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

type LivenessState struct {
	Status string `json:"status"`
}

type ReadinessState struct {
	Status string `json:"status"`
}

type HealthComponent struct {
	Liveness  LivenessState  `json:"livenessState"`
	Readiness ReadinessState `json:"readinessState"`
}

type HealthResponse struct {
	Status     string          `json:"status"`
	Components HealthComponent `json:"components"`
	Groups     []string        `json:"groups"`
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
var inComMCU_ bool

// Error code
const SUCCESS = 0
const ERROR_POWER_ONOFF = 1
const ERROR_UNKNOWN_CMD = 2
const ERROR_WRITING = 100
const ERROR_NO_PORT_FOUND = 101
const ERROR_OPEN_PORT = 102
const ERROR_RESET_OUTBUFFER = 103
const ERROR_RESET_INBUFFER = 104
const ERROR_PORT_NOT_SPECIFIED = 105
const ERROR_PORT_BUSY = 200
const ERROR_READING = 201
const ERROR_NO_DATA_READ = 202

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
		//return
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

	// Set swagger info
	docs.SwaggerInfo.Title = "Infra-External API"
	docs.SwaggerInfo.Description = "This is a power-controller backend server"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1/infra-external/power"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	//
	inComMCU_ = false

	router := gin.Default()

	setupSwagger(router)

	basePath := "/api/v1/infra-external/power"

	router.GET(basePath+"/set/:id/:cmd", setPower)
	router.GET(basePath+"/initialize", initialize)
	router.GET(basePath+"/get/:id", getPower)

	router.GET("/actuator/health", healthCheck)

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

func zeroPad(number int) string {
	tmpStr := strconv.FormatInt(int64(number), 10)
	return fmt.Sprintf("%04s", tmpStr)
}

func setupSwagger(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func healthCheck(c *gin.Context) {
	var healthResponse HealthResponse

	healthResponse.Status = "UP"
	healthResponse.Components.Liveness.Status = "UP"
	healthResponse.Components.Readiness.Status = "UP"
	healthResponse.Groups = append(healthResponse.Groups, "liveness")
	healthResponse.Groups = append(healthResponse.Groups, "readiness")
	c.IndentedJSON(http.StatusOK, healthResponse)
}

// setPower godoc
// @Summary      Power on/off workstation
// @Description  Power on/off workstation with id
// @Tags         infra-external
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Workstation ID in power-controller"
// @Param 		 cmd  path		string true "Power controll command(S:power-on, Q:shutdown(OS), E:power-off(HW)"
// @Success      200  {object}  string "Successfully power on/off workstation with the given id"
// @Router       /set/{id}/{cmd} [get]
func setPower(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := string(chars)

	paramId := c.Param("id")
	paramCmd := c.Param("cmd")

	tmpParamId, _ := strconv.Atoi(string(paramId))
	tmpCmd := paramCmd + zeroPad(tmpParamId)
	//logger.Infof("Sent command : %v", tmpCmd)

	code, err := pwCtrl.setCommand(tmpCmd, &mesg, 100)
	if err != nil {
		logger.Info(err.Error())
		if code == ERROR_PORT_BUSY {
			var failResponse McuResponseFail

			failResponse.State = "fail"
			failResponse.Message = err.Error()
			failResponse.ErrorType = strconv.Itoa(code)
			c.IndentedJSON(http.StatusRequestTimeout, failResponse)
			return
		}
	}
	//logger.Infof("MCU response : %v", mesg)

	var tmpResponse CmdResult
	tmpResponse.Cmd = tmpCmd
	tmpResponse.Res = mesg

	mcuCode := -1
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
		failResponse.Message = err.Error()
		failResponse.ErrorType = strconv.Itoa(code)
		if mcuCode == 9 {
			c.IndentedJSON(http.StatusBadRequest, failResponse)
		} else {
			c.IndentedJSON(http.StatusInternalServerError, failResponse)
		}
	}
}

// getPower godoc
// @Summary      Check power state of worktation
// @Description  Check power state with workstation id
// @Tags         infra-external
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Workstation ID in power-controller"
// @Success      200  {object}  string "Power state(on/ff) identified"
// @Router       /get/{id} [get]
func getPower(c *gin.Context) {
	chars := make([]byte, 64)
	mesg := string(chars)

	paramId := c.Param("id")

	tmpParamId, _ := strconv.Atoi(string(paramId))
	tmpCmd := "C" + zeroPad(tmpParamId)
	//logger.Infof("Sent command : %v", tmpCmd)

	code, err := pwCtrl.setCommand(tmpCmd, &mesg, 100)
	//logger.Infof("MCU response : %v", mesg)
	if err != nil {
		logger.Info(err.Error())

		if code == ERROR_PORT_BUSY {
			var failResponse McuResponseFail
			failResponse.State = "fail"
			failResponse.Message = err.Error()
			failResponse.ErrorType = strconv.Itoa(code)
			c.IndentedJSON(http.StatusRequestTimeout, failResponse)
			return
		} else if code == ERROR_PORT_NOT_SPECIFIED {
			var failResponse McuResponseFail

			failResponse.State = "fail"
			failResponse.Message = err.Error()
			failResponse.ErrorType = strconv.Itoa(code)
			c.IndentedJSON(http.StatusInternalServerError, failResponse)
			return
		}
	}

	//logger.Info()

	var tmpResponse CmdResult
	tmpResponse.Cmd = tmpCmd
	tmpResponse.Res = mesg
	//logger.Info("RESPONSE :", mesg, len(mesg), len(tmpResponse.Res))

	mcuCode := 0
	if len(tmpResponse.Res) > 0 {
		mcuCode, _ = strconv.Atoi(tmpResponse.Res[:1])
	} else {
		var failResponse McuResponseFail
		failResponse.State = "fail"
		failResponse.Message = err.Error()
		failResponse.ErrorType = strconv.Itoa(code)
		c.IndentedJSON(http.StatusInternalServerError, failResponse)
		//httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	//logger.Info("mcuCode=", mcuCode)

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
		failResponse.Message = err.Error()
		failResponse.ErrorType = strconv.Itoa(code)
		if mcuCode == 9 {
			c.IndentedJSON(http.StatusBadRequest, failResponse)
		} else {
			c.IndentedJSON(http.StatusInternalServerError, failResponse)
		}
	}
}

// intialize godoc
// @Summary      Initialize serial port
// @Description  Initialize serial port
// @Tags         infra-external
// @Accept       json
// @Produce      json
// @Success      200  {object}  string "Successfully initialized the serial port"
// @Router       /initialize [get]
func initialize(c *gin.Context) {
	code, err := pwCtrl.intializeConnection()
	if err != nil {
		logger.Info("Failed to initialize serial port")
	} else {
		logger.Infof("Successfully initialized serial port : %v", pwCtrl.portName)
	}

	if err != nil {
		var failResponse McuResponseFail
		failResponse.State = "fail"
		failResponse.Message = err.Error()
		failResponse.ErrorType = strconv.Itoa(code)
		c.IndentedJSON(http.StatusInternalServerError, failResponse)
	} else {
		var response McuResponse
		response.Data = true
		response.State = "success"
		response.ElapsedSeconds = 0
		c.IndentedJSON(http.StatusOK, response)
	}
}

func (pwctl *PwCtrl) setCommand(cmdStr string, response *string, sleepUTime int) (int, error) {
	if inComMCU_ {
		error := errors.New("Busy serial communication")
		//logger.Info(error)
		return ERROR_PORT_BUSY, error
	} else {
		inComMCU_ = true
	}

	// NOTE : Clear input buffer before writing
	if pwctl.serialPort != nil {
		if pwctl.serialPort.ResetInputBuffer() != nil {
			inComMCU_ = false
			return ERROR_RESET_INBUFFER, errors.New("Failed to reset input buffer")
		}
	} else {
		inComMCU_ = false
		return ERROR_PORT_NOT_SPECIFIED, errors.New("Serial port not specified")
	}

	logger.Info("Sent command : ", cmdStr)

	err := pwctl.write([]byte(cmdStr))
	if err != nil {
		//Re-initializing as a separate thread
		if !pwctl.reIntializing {
			logger.Info("Re-initializing serial port")
			go pwctl.reIntializeConnection()
		}
		logger.Info(err.Error())

		inComMCU_ = false
		return ERROR_WRITING, err
	} else {
		// NOTE : Some sleep before reading to avoid dropping in response
		time.Sleep(200 * time.Millisecond)

		tmpRes := make([]byte, 64)
		n, err := pwctl.read(tmpRes)
		if n == 0 {
			errMesg := "ERROR : no data read"
			logger.Info(errMesg)
			inComMCU_ = false
			return ERROR_NO_DATA_READ, errors.New(errMesg)
		}

		if err != nil {
			logger.Info(err.Error())

			inComMCU_ = false
			return ERROR_READING, err
		}

		*response = string(tmpRes)
		logger.Info("Received data : ", (*response)[:1])
		//fmt.Println("n=", n)
		//fmt.Println("response = ", response[:n])

		if (*response)[n-1] != '\n' {
			logger.Info("WARNING : no newline character in response")
			inComMCU_ = false
			return SUCCESS, nil
		} else if (*response)[0] == '9' {
			errMesg := "ERROR : unknown command or wrong rack-number"
			logger.Info(errMesg)
			inComMCU_ = false
			return ERROR_UNKNOWN_CMD, errors.New(errMesg)
		}
		//---------------------------------------------------------
		// NOTE : Ignore CODE=8
		// because in many cases the power state is not
		// correctly sent right after executing a power on/off command.
		//---------------------------------------------------------
		//else if response[0] == '8' {
		//	errMesg = "ERROR : failed to power on/off"
		//	logger.Info(errMesg)
		//	return ERROR_POWER_ONOFF, errors.New(errMesg)
		//}

		inComMCU_ = false
		return SUCCESS, nil
	}
}

func (pwctl *PwCtrl) reIntializeConnection() {
	// To prevent multiple executions of re-initializing
	if !pwctl.reIntializing {
		pwctl.reIntializing = true
		inComMCU_ = true

		for {
			_, err := pwctl.intializeConnection()
			if err == nil {
				pwctl.reIntializing = false
				inComMCU_ = false
				break
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func (pwctl *PwCtrl) read(buff []byte) (int, error) {
	if pwctl.serialPort == nil {
		return 0, errors.New("serial port not initialized")
	}

	n, err := pwctl.serialPort.Read(buff)
	if err != nil {
		return 0, err
	}
	//fmt.Println("Data received : size = ", n)
	//fmt.Println(buff[:n])
	if n == 0 {
		return 0, errors.New("no data received or timeout")
	}
	return n, nil
}

func (pwctl *PwCtrl) write(data []byte) error {
	if pwctl.serialPort == nil {
		return errors.New("serial port not initialized")
	}
	_, err := pwctl.serialPort.Write(data)
	if err != nil {
		return err
	}

	//fmt.Println("- Write data : size = ", n)
	return nil
}

func (pwctl *PwCtrl) intializeConnection() (int, error) {
	if pwctl.serialPort != nil {
		pwctl.serialPort.Close()
	}

	err := pwctl.findSerialPort()
	if err != nil {
		if !pwctl.reIntializing {
			logger.Info(err.Error())
		}
		return ERROR_NO_PORT_FOUND, err
	}

	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	pwctl.serialPort, err = serial.Open(pwctl.portName, mode)
	if err != nil {
		logger.Info(err.Error())
		return ERROR_OPEN_PORT, err
	}

	pwctl.serialPort.SetReadTimeout(time.Duration(pwctl.readTimeOut) * time.Second)

	err = pwctl.serialPort.ResetInputBuffer()
	if err != nil {
		logger.Info(err.Error())
		return ERROR_RESET_INBUFFER, err
	}

	err = pwctl.serialPort.ResetOutputBuffer()
	if err != nil {
		logger.Info(err.Error())
		return ERROR_RESET_OUTBUFFER, err
	}

	logger.Info("Serial port re-initialized : ", pwctl.portName)

	return SUCCESS, nil
}

func (p *PwCtrl) findSerialPort() error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return err
	}

	portList := make([]string, 0)

	for _, port := range ports {
		if len(port) >= len(p.portPrefix) && port[:len(p.portPrefix)] == p.portPrefix {
			portList = append(portList, port)
		}
	}

	if len(portList) != 1 {
		return errors.New("no or multiple ports found")
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
