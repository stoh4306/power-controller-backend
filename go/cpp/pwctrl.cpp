#include "pwctrl.h"

#include <string>
#include <iostream>

// C library headers
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Linux headers
#include <fcntl.h> // Contains file controls like O_RDWR
#include <errno.h> // Error integer and strerror() function
#include <termios.h> // Contains POSIX terminal control definitions
#include <unistd.h> // write(), read(), close()

#include <dirent.h> // To find serial port
#include <thread>
#include <chrono>

void PwCtrlBackend::setPortNamePrefix(std::string prefix)
{
    portNamePrefix_ = prefix;
    //std::cout << "prefix=" << portNamePrefix_ << std::endl;
}

void PwCtrlBackend::setMinimumBytes(int minByte)
{
    minimumBytesToRead_ = minByte;
}

void PwCtrlBackend::setMaxReadTime(int deciSec)
{
    maxReadTime_ = deciSec;
}

int PwCtrlBackend::find_serial_port(std::string& portName)
{
    std::string fileName;

    DIR * dir;
    struct dirent *ent;
    if ((dir = opendir("/dev")) != NULL)
    {
        bool portFound = false;
        while((ent=readdir(dir)) != NULL)
        {
            fileName = ent->d_name;
            if (fileName.substr(0,6) == portNamePrefix_)
            {
                portName = std::string("/dev/")+fileName;
                portFound = true;
                break;
            }
        }

        if (portFound) return SUCCESS; 
        else return ERR_PORT_NOT_FOUND; 
    }
    else
    {
        if (debugging_) std::cerr << "ERROR in finding serial port : failed to opendir(/dev)" << std::endl;
        return ERR_FINDING_SERIAL_PORT;
    }
}

bool PwCtrlBackend::open_serial_port(int* port, const char* portName)
{
    *port =  open(portName, O_RDWR);
    if (debugging_)     std::cout << "port=" << *port << std::endl;

    if (*port > 0)   return true; 
    else return false;
}

int PwCtrlBackend::configure_serial_port(int* serial_port)
{
    // Create new termios struct, we call it 'tty' for convention
    struct termios tty;

    // Read in existing settings, and handle any error
    if(tcgetattr(*serial_port, &tty) != 0) {
        if (debugging_)     printf("Error %i from tcgetattr: %s\n", errno, strerror(errno));
        return 1;
    }

    tty.c_cflag &= ~PARENB; // Clear parity bit, disabling parity (most common)
    tty.c_cflag &= ~CSTOPB; // Clear stop field, only one stop bit used in communication (most common)
    tty.c_cflag &= ~CSIZE; // Clear all bits that set the data size 
    tty.c_cflag |= CS8; // 8 bits per byte (most common)
    tty.c_cflag &= ~CRTSCTS; // Disable RTS/CTS hardware flow control (most common)
    tty.c_cflag |= CREAD | CLOCAL; // Turn on READ & ignore ctrl lines (CLOCAL = 1)

    tty.c_lflag &= ~ICANON;
    tty.c_lflag &= ~ECHO; // Disable echo
    tty.c_lflag &= ~ECHOE; // Disable erasure
    tty.c_lflag &= ~ECHONL; // Disable new-line echo
    tty.c_lflag &= ~ISIG; // Disable interpretation of INTR, QUIT and SUSP
    tty.c_iflag &= ~(IXON | IXOFF | IXANY); // Turn off s/w flow ctrl
    tty.c_iflag &= ~(IGNBRK|BRKINT|PARMRK|ISTRIP|INLCR|IGNCR|ICRNL); // Disable any special handling of received bytes

    tty.c_oflag &= ~OPOST; // Prevent special interpretation of output bytes (e.g. newline chars)
    tty.c_oflag &= ~ONLCR; // Prevent conversion of newline to carriage return/line feed
    // tty.c_oflag &= ~OXTABS; // Prevent conversion of tabs to spaces (NOT PRESENT ON LINUX)
    // tty.c_oflag &= ~ONOEOT; // Prevent removal of C-d chars (0x004) in output (NOT PRESENT ON LINUX)

    tty.c_cc[VTIME] = maxReadTime_;    // Wait for up to VTIME deciseconds, returning as soon as any data is received.
    tty.c_cc[VMIN] = minimumBytesToRead_;      // Minimum bytes => 3
    if (debugging_) std::clog << "MaxReadTime=" << maxReadTime_/10 << " sec" << "\n"
        << "Minimum Bytes=" << minimumBytesToRead_ << std::endl;

    // Set in/out baud rate to be 9600
    cfsetispeed(&tty, B9600);
    cfsetospeed(&tty, B9600);

    // Save tty settings, also checking for error
    if (tcsetattr(*serial_port, TCSANOW, &tty) != 0) {
        if (debugging_) printf("Error %i from tcsetattr: %s\n", errno, strerror(errno));
        return 2;
    }

    return 0;
}

int PwCtrlBackend::clearSerialIOBuffer()
{
    // Clear serial port buffer
    int result = tcflush(serial_port_, TCIOFLUSH);
    if ( result < 0 )
    {
        if (debugging_) std::clog << "Warning, failed to clear serial I/O buffer" << std::endl;
        return 1;
    }

    return 0;
}

int PwCtrlBackend::batch_init_connection()
{
    int initResult = 0;
    while(1)
    {
        if (should_run_ == false) break;

        if (reInitRequired_ == true)
        {
            initResult = initialize_connection();
            if ( initResult == SUCCESS )   
            {
                std::cout << "[SERIAL-COM] Successfully initialize serial port : " 
                    << portName_ << std::endl;

                reInitRequired_ = false;
            }
        }
        sleep(reconnectIntervalInSec_);
    }

    return SUCCESS;
}

int PwCtrlBackend::initialize_connection()
{
    std::string newPortName;
    int result = find_serial_port(newPortName);
    if ( result > 0 ) return result;

    if (debugging_)
        std::clog << "Port found : " << newPortName << std::endl;
    
    if ( portName_.length() == 0 )
    {
        portName_ = newPortName;
    }

    int old_serial_port = 0;
    if (serial_port_ > 0)
    {
        old_serial_port = serial_port_;
        result = close(serial_port_);
        if (debugging_) std::clog << "close port result = " << result << std::endl;
    }

    if (open_serial_port(&serial_port_, newPortName.c_str()) == false)
    {
        if (debugging_) std::clog << "Error, can't open serial port : " << newPortName << std::endl;
        serial_port_ = old_serial_port;
        return ERR_OPEN_PORT;
    }
    else 
    {
        portName_ = newPortName;

        // Configure serial port
        result = configure_serial_port(&serial_port_);
        if ( result != 0 )
        {
            if (debugging_) std::clog << "Error, failed to configure the serial port" << std::endl;
            return ERR_CONFIGURE_PORT;
        }

        if (debugging_) std::clog << "configure port done" << std::endl;

        // Clear serial buffer
        result = clearSerialIOBuffer();
        if (result != 0)
        {
            if (debugging_) std::clog << "Error, failed to clear serial buffer : " 
                << serial_port_ << " " << portName_ << std::endl; 
            return ERR_CLEAR_SERIAL_BUFFER;
        }
        if (debugging_) std::clog << "clear serial buffer done" << std::endl;

        initialized_ = true;
        reInitRequired_ = false;
        
        return SUCCESS;
    } 
}

int PwCtrlBackend::writeSerialPort(std::string mesg)
{
    if (mesg[mesg.length()-1] == '\n')
    {
        // NOTE : This is for cgo. Not needed in c/c++
        mesg[mesg.length()-1] = '\r';
        if (debugging_) std::clog << "line feed deleted in command string" << std::endl;
    }
    else
    {
        mesg = mesg + "\r";
    }

    //std::cout << mesg.length() << " " << mesg << std::endl;
    //for (int i = 0; i < mesg.length(); ++i)
    //    std::cout << "[" << (unsigned int)mesg[i] << "] ";
    //std::cout << std::endl;

    int result = write(serial_port_, mesg.c_str(), mesg.length());
    if (result < 0)
    {
        if (debugging_) std::clog << "Error in writing : " << result << std::endl;
        return ERR_WRITE_SERIAL_PORT;
    }
    return SUCCESS;
}

int PwCtrlBackend::readSerialPort(std::string& mesg)
{
    memset(read_buf_, '\0', sizeof(read_buf_));

    // Read bytes. The behaviour of read() (e.g. does it block?,
    // how long does it block for?) depends on the configuration    
    // settings above, specifically VMIN and VTIME
    int num_bytes = read(serial_port_, &read_buf_, sizeof(read_buf_));

    // n is the number of bytes read. n may be 0 if no bytes were received, and can also be -1 to signal an error.
    if (num_bytes < 0) {
        printf("[SERIAL-COM] Error reading: %s", strerror(errno));
        return ERR_READ_SERIAL_PORT;
    }

    // Here we assume we received ASCII data, but you might be sending raw bytes (in that case, don't try and
    // print it to the screen like this!)
    if (debugging_) 
    {
        printf("[SERIAL-COM] Read %i bytes. ", num_bytes );
        if (num_bytes > 0) 
            printf("Received message: %c\n", read_buf_[0]);
        else
            printf("\n");
    }
    //for (int i = 0; i < num_bytes; ++i )
    //{
    //    std::cout << "[" << (unsigned int)read_buf_[i] << "] ";
    //}
    //std::cout << std::endl;
    mesg = std::string(read_buf_);
    //*/

    return SUCCESS;
}

int PwCtrlBackend::set_command(std::string cmdStr, std::string& response, int sleepUTime)
{
    int result = 0;
    result = writeSerialPort(cmdStr);

    if (result != 0)
    {
        initialized_ = false;

        reInitRequired_ = true;
        /*int initResult = 0;
        while(1)
        {
            if (initialized_) break;

            initResult = initialize_connection();
            if ( initResult == SUCCESS )   
            {
                std::cout << "[SERIAL-COM] Successfully initialize serial port : " 
                    << portName_ << std::endl;
                break;
            }
            sleep(reconnectIntervalInSec_);
        }//*/
        return result;
    } 
    else
    {
        result = readSerialPort(response);
        if (result != SUCCESS) return result;

        if (response.length() == 0)
        {
            // Sending command was ok, but failed to get response from MCU
            result = ERR_NO_RESPONSE;
            std::clog << "[SERIAL-COM] ERROR, no response from MCU" << std::endl;
        }
        else if (*(response.end()-1) != '\n')
        {
            result = ERR_INCOMPLETE_RESPONSE;
            std::clog << "[SERIAL-COM] Warning, incomplete response. It doesn't end with the line feed" << std::endl;
        }
        else if (response[0] == '8')
        {
            result = ERR_FAIL_POWER_ONOFF;
            std::clog << "[SERIAL-COM] ERROR, failed to power on/off" << std::endl;
        }
        else if (response[0] == '9')
        {
            result = ERR_UNKNOWN_CMD_WRONG_RACKNUM;
            std::clog << "[SERIAL-COM] ERROR, unknown command or wrong rack-number" << std::endl;
        }
        return result;
    }

}

int PwCtrlBackend::closePort()
{
    close(serial_port_);
    serial_port_ = 0;
    portName_ = "";
    memset(read_buf_, '\0', sizeof(read_buf_));

    return 0;
}

int PwCtrlBackend::setDebuggingMode(int mode)
{
    if (mode == 0) debugging_ = false; else debugging_ = true;
    return 0;
}

int PwCtrlBackend::startInitThread()
{
    initConnThread_ = std::thread(&PwCtrlBackend::batch_init_connection, this);
    return 0;
}

int PwCtrlBackend::stopInitThread()
{
    if (initConnThread_.joinable()) {
        should_run_ = false; // Signal the thread to stop
        initConnThread_.join();  // Wait for the thread to finish
    }
    return 0;
}

PwCtrlBackend::~PwCtrlBackend()
{
    stopInitThread();

    if (closePort() == 0)
    {
        if (debugging_) std::clog << "[SERIAL-COM] PwCtrlBackend successfully destroyed" << std::endl;
    }
    else
    {
        if (debugging_) std::clog << "[SERIAL-COM] ERROR in destructor::closePort()" << std::endl;
    }
}

PwCtrlBackend::PwCtrlBackend()
{
    serial_port_ = 0;
    portName_ = "";
    portNamePrefix_ = "";
    char read_buf_[256];
    maxReadTime_ = 0;
    minimumBytesToRead_ = 0;
    reconnectIntervalInSec_ = 5;
    isReconnecting_ = false;
    initialized_ = false;
    reInitRequired_ = true;
    debugging_ = false;

    should_run_ = true;
}