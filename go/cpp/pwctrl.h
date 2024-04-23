#ifndef PWCTRL_BE_H
#define PWCTRL_BE_H

#include <string>

#define SUCCESS                 0
#define ERR_OPEN_PORT           1
#define ERR_CONFIGURE_PORT      2
#define ERR_CLEAR_SERIAL_BUFFER 3
#define ERR_WRITE_SERIAL_PORT   4
#define ERR_READ_SERIAL_PORT    5
#define ERR_FINDING_SERIAL_PORT 6
#define ERR_PORT_NOT_FOUND      7
#define ERR_NO_RESPONSE         8
#define ERR_INCOMPLETE_RESPONSE 9
#define ERR_FAIL_POWER_ONOFF    10
#define ERR_UNKNOWN_CMD_WRONG_RACKNUM 11

class PwCtrlBackend
{
public:
    PwCtrlBackend();
    ~PwCtrlBackend();

    void setPortNamePrefix(std::string prefix);
    void setMinimumBytes(int minByte);
    void setMaxReadTime(int deciSec);
    int find_serial_port(std::string& portName);
    bool open_serial_port(int* port, const char* portName);
    int configure_serial_port(int* serial_port);
    int clearSerialIOBuffer();
    int initialize_connection();
    int writeSerialPort(std::string mesg);
    int readSerialPort(std::string& mesg);
    int set_command(std::string cmdStr, std::string& response, int sleepUTime);
    int closePort();

public:
    int serial_port_ = 0;
    std::string portName_ = "";
    std::string portNamePrefix_;
    char read_buf_[256];
    int maxReadTime_ = 0;
    int minimumBytesToRead_ = 0;
    int reconnectIntervalInSec_ = 5;
    bool isReconnecting_ = false;
};

#endif