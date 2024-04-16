#include "pwctrl-wrapper.h"
#include "pwctrl.h"
#include <cstring>
#include <iostream>

int setPortNamePrefix(void* pwCtrlBe, const char* prefix)
{
    if (pwCtrlBe)
    {
        ((PwCtrlBackend*)pwCtrlBe)->setPortNamePrefix(std::string(prefix));
        return SUCCESS;
    } 
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    } 
}

int setMinimumBytes(void* pwCtrlBe, int minBytes)
{
    if (pwCtrlBe)
    {
        ((PwCtrlBackend*)pwCtrlBe)->setMinimumBytes(minBytes);
        return SUCCESS;
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int setMaxReadTime(void* pwCtrlBe, int deciSec)
{
    if (pwCtrlBe)
    {
        ((PwCtrlBackend*)pwCtrlBe)->setMaxReadTime(deciSec);
        return SUCCESS;
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int find_serial_port(void* pwCtrlBe, char* portName, int n)
{
    if (pwCtrlBe)
    {
        std::string portNameFound;
        int result = ((PwCtrlBackend*)pwCtrlBe)->find_serial_port(portNameFound);

        if ( portNameFound.length()+1 <= n )
        {
            //strcpy(response, tmpResponse.c_str());
            memcpy(portName, portNameFound.data(), portNameFound.length());
            portName[portNameFound.length()] = '\0';
        }
        else
        {
            std::cerr << "Warning, size of received mesg exceeds the buffer size" << std::endl;
            memcpy(portName, portNameFound.data(), n-1);
            portName[n-1] = '\0';
        }

        return result;
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int open_serial_port(void* pwCtrlBe, int* port, const char* portName)
{
    if (pwCtrlBe)
    {
        bool result = ((PwCtrlBackend*)pwCtrlBe)->open_serial_port(port, portName);

        if (result == false)    
            return ERR_OPEN_PORT;
        else
            return SUCCESS;
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int configure_serial_port(void* pwCtrlBe, int* serial_port)
{
    if (pwCtrlBe)
    {
        return ((PwCtrlBackend*)pwCtrlBe)->configure_serial_port(serial_port);
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int clearSerialIOBuffer(void* pwCtrlBe)
{
    if (pwCtrlBe)
    {
        return ((PwCtrlBackend*)pwCtrlBe)->clearSerialIOBuffer();
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int initialize_connection(void* pwCtrlBe)
{
    if (pwCtrlBe)
    {
        return ((PwCtrlBackend*)pwCtrlBe)->initialize_connection();
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int writeSerialPort(void* pwCtrlBe, const char* mesg)
{
    if (pwCtrlBe)
    {
        return ((PwCtrlBackend*)pwCtrlBe)->writeSerialPort(std::string(mesg));
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int readSerialPort(void* pwCtrlBe, char* mesg, int n)
{
    if (pwCtrlBe)
    {
        std::string receivedMesg;
        int result = ((PwCtrlBackend*)pwCtrlBe)->readSerialPort(receivedMesg);

        if ( receivedMesg.length()+1 <= n )
        {
            //strcpy(response, tmpResponse.c_str());
            memcpy(mesg, receivedMesg.data(), receivedMesg.length());
            mesg[receivedMesg.length()] = '\0';
        }
        else
        {
            std::cerr << "Warning, size of received mesg exceeds the buffer size" << std::endl;
            memcpy(mesg, receivedMesg.data(), n-1);
            mesg[n-1] = '\0';
        }

        return result;
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}

int set_command(void* pwCtrlBe, const char* cmdStr, char* response, int n, int sleepUTime)
{
    if (pwCtrlBe)
    {
        std::string tmpResponse;
        int result = ((PwCtrlBackend*)pwCtrlBe)->set_command(std::string(cmdStr), tmpResponse, 1000);
        std::cout << "result=" << result << std::endl;
        //std::cout << "response ptr=" << response << std::endl;
        //std::cout << "response array size=" << sizeof(response) << std::endl;
        //std::cout << "tmpResponse array size=" << sizeof(tmpResponse) << std::endl;

        if ( tmpResponse.length()+1 <= n )
        {
            //strcpy(response, tmpResponse.c_str());
            memcpy(response, tmpResponse.data(), tmpResponse.length());
            response[tmpResponse.length()] = '\0';
        }
        else
        {
            std::cerr << "Warning, size of received mesg exceeds the buffer size" << std::endl;
            memcpy(response, tmpResponse.data(), n-1);
            response[n-1] = '\0';
        }

        return result;
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}
int closePort(void* pwCtrlBe)
{
    if (pwCtrlBe)
    {
        return ((PwCtrlBackend*)pwCtrlBe)->closePort();
    }
    else
    {
        return ERR_NULL_PWCTRL_PTR;
    }
}
void* createPwctrlBackend()
{
    PwCtrlBackend * pwCtrlBe = new PwCtrlBackend();
    return (void*)pwCtrlBe;
}

void deletePwctrlBackend(void* pwCtrlBe)
{
    delete (PwCtrlBackend*)pwCtrlBe;
    pwCtrlBe = NULL;
}