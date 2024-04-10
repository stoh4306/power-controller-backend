#include "pwctl.h"

int main(int argc, char** argv) 
{
    std::string portName = argv[1];
    int sleepUTime = atoi(argv[2]);

    int result = initialize_connection();
    if (result > 0)
    {
        return result;
    }

    std::string mesg;

    while(true)
    {
        std::cout << "CMD : ";
        std::string cmdMesg;
        std::cin >> cmdMesg;

        if (cmdMesg.substr(0,1) == "r")
        {
            readSerialPort(mesg);
            
        }
        else if (cmdMesg.substr(0,1) == "w")
        {
            set_command(cmdMesg.substr(1,cmdMesg.length()), mesg, sleepUTime);
        }
        else if (cmdMesg.substr(0,1) == "c")
        {
            clearSerialIOBuffer();
        }
        else if (cmdMesg.substr(0,1) == "i")
        {
            //std::string newPortName = cmdMesg.substr(1,cmdMesg.length());
            //std::cout << newPortName << std::endl;
            //initialize_connection(newPortName);
            initialize_connection();
        }
        else if (cmdMesg.substr(0,1) == "q")
        {
            closePort();
            break;
        }
        else
        {
            std::cout << "Unkown command : " << cmdMesg[0] << std::endl;
        }
    }  
}