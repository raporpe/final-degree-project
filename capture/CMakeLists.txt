cmake_minimum_required(VERSION 3.18)

# set the project name
project(scanner)

#add_library(sqlite3 STATIC "include/sqlite3/sqlite3.c")
include_directories(scanner "include")

# Add the sqlite wrapper
add_subdirectory("include/SQLiteCpp")
add_subdirectory("include/libtins")
add_subdirectory("include/fmt")

# add the executable
add_executable(scanner "src/main.cpp" "src/helpers.cpp")
#find_package(Threads REQUIRED)
target_link_libraries(scanner PRIVATE pthread tins curl sqlite3 dl SQLiteCpp fmt libtins)
SET(CMAKE_CXX_FLAGS  "${CMAKE_CXX_FLAGS} -Wno-psabi")
