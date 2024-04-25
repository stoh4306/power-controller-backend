#ifndef PWCTRL_BE_WRAPPER_H
#define PWCTRL_BE_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

#define ERR_NULL_PWCTRL_PTR         100
#define ERR_SET_PORTNAME_PREFIX     101
#define ERR_SET_MIN_BYTES           102
#define ERR_SET_MAX_READ_TIME       103
#define ERR_COPY_CHAR_ARRAY         200

int setPortNamePrefix(void* pwCtrlBe, const char* prefix);
int setMinimumBytes(void* pwCtrlBe, int minBytes);
int setMaxReadTime(void* pwCtrlBe, int deciSec);
int find_serial_port(void* pwCtrlBe, char* portName, int n);
int open_serial_port(void* pwCtrlBe, int* port, const char* portName);
int configure_serial_port(void* pwCtrlBe, int* serial_port);
int clearSerialIOBuffer(void* pwCtrlBe);
int initialize_connection(void* pwCtrlBe, int maxLengthPortName, char* portName);
int writeSerialPort(void* pwCtrlBe, const char* mesg);
int readSerialPort(void* pwCtrlBe, char* mesg, int n);
int set_command(void* pwCtrlBe, const char* cmdStr, char* response, int n, int sleepUTime);
int closePort(void* pwCtrlBe);
int setDebuggingMode(void* pwCtrlBe, int mode);
int getPortName(void* pwCtrlBe, int maxLength, char* portName);
void* createPwctrlBackend();
void  deletePwctrlBackend(void* pwCtrlBe);

#ifdef __cplusplus
}
#endif

#endif