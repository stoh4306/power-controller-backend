#include "pwctl.h"

int main(int argc, char** argv) 
{
    if (argc < 4)
    {
        std::cout << "*******************" << "\n"
            << " Power-Controller-Backend" << "\n"
            << " - usage : pwctl <arg1> <arg2> <arg3>" << "\n"
            << "  . arg1 : port name prefix (ex: ttyACM or ttyUSB)" << "\n"
            << "  . arg2 : max. reading time in deciseconds (10decisec = 1sec)" << "\n"
            << "  . arg3 : minimum bytes to read" << "\n"
            << "  . (example) pwctl ttyACM 100 0" << "\n"
            << std::endl;
        return 1;
    }

    std::string portNamePrefix = argv[1];
    int maxReadTime = atoi(argv[2]);
    int minByte = atoi(argv[3]);

    setPortNamePrefix(portNamePrefix);

    setMaxReadTime(maxReadTime);
    setMinimumBytes(minByte);

    int sleepUTime = 1000000; // 1 sec

    int result = initialize_connection();
    if (result > 0)
    {
        std::cerr << "ERROR, failed to initialize connection " << std::endl;
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
            result = set_command(cmdMesg.substr(1,cmdMesg.length()), mesg, sleepUTime);
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