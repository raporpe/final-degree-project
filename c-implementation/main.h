#ifndef MAIN_H
#define MAIN_H

#include <string>
#include <tins/tins.h>
#include <vector>
#include <bitset>

using namespace std;
using namespace Tins;

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

const int WINDOW_SIZE = 15;
const int FRAME_TIME = 60;
const float ACTIVITY_PERCENTAGE = 0.6;

class PacketManager {
   private:
    vector<Dot11ProbeResponse> probeResponses;
    vector<Dot11ProbeRequest> probeRequests;
    int currentSecond = 0;
    map<mac, bitset<WINDOW_SIZE>> store;
    bool uploadBackend = false;

    int getCurrentTime();

    void uploadToBackend();

    void checkTimeIncrease();

    void addAndTickMac(mac mac_address);

    int getActiveDevices();

    void tickMac(mac mac_address);

   public:
    PacketManager(char *upload_backend);

    void registerProbeRequest(Dot11ProbeRequest *frame);

    void registerProbeResponse(Dot11ProbeResponse *frame);

    void registerControl(Dot11Control *frame);
};

#endif