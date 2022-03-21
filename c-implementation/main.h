#ifndef MAIN_H
#define MAIN_H

#include <tins/tins.h>

#include <bitset>
#include <string>
#include <vector>

using namespace std;
using namespace Tins;

const int WINDOW_SIZE = 15;
const int FRAME_TIME = 60;
const float ACTIVITY_PERCENTAGE = 0.6;

typedef HWAddress<6> mac;

struct ProbeRequest {
    string stationMac;
    string intent;
    string time;
    int frequency;
    int power;
};

struct UploadJSONData {
    string deviceID;
    std::vector<ProbeRequest> probeRequests;
};

struct StoreObject {
    bitset<WINDOW_SIZE> state;
    int signal_strength;
};



class PacketManager {
   private:
    vector<Dot11ProbeResponse> probeResponses;
    vector<Dot11ProbeRequest> probeRequests;
    int currentSecond = 0;
    map<mac, StoreObject> store;
    bool uploadBackend = false;
    char* device_id;

    void uploadToBackend();

    void checkTimeIncrease();

    void addAndTickMac(mac mac_address, int signal_strength);

    void tickMac(mac mac_address, int signal_strength);

    int getActiveDevices();

   public:
    PacketManager(char *upload_backend, char* device_id);

    void registerProbeRequest(Dot11ProbeRequest *frame);

    void registerProbeResponse(Dot11ProbeResponse *frame);

    void registerControl(Dot11Control *frame);

    void registerData(Dot11Data *frame);

};

#endif