#include "helpers.h"
#include "main.h"

#include <chrono>
#include <tins/tins.h>

using namespace std::chrono;
    
int getCurrentTime() {
    return duration_cast<seconds>(system_clock::now().time_since_epoch())
        .count();
}


mac getStationMAC(Tins::Dot11Data *frame) {
    bool from = frame->from_ds();
    bool to = frame->to_ds();


    if(!to && !from){
        return frame->addr2();
    } else if (!to && from) {
        return frame->addr1();
    } else if (to && !from) {
        return frame->addr2();
    } else {
        return mac(nullptr);
    }
}

bool isMacValid(mac address) {
    return true;
}

bool isMacFake(mac address) {
    return (address[0] & 0x02) == 0x02;
}