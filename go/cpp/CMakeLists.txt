cmake_minimum_required(VERSION 3.11)

project(pwctrl)

#set(CMAKE_CXX_STANDARD 11)
#set(CMAKE_CXX_STANDARD_REQUIRED ON)
#set(CMAKE_CXX_EXTENSIONS OFF)

find_package(Threads REQUIRED)

add_library(pwctrlbe SHARED pwctrl.cpp pwctrl-wrapper.cpp)

add_executable(pwctrl pwctrl-main.cpp)
target_link_libraries(pwctrl PRIVATE pwctrlbe ${CMAKE_THREAD_LIBS_INIT})