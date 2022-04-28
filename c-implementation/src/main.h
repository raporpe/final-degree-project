#ifndef MAIN_H
#define MAIN_H

#include <SQLiteCpp/SQLiteCpp.h>
#include <tins/tins.h>

#include <bitset>
#include <string>
#include <vector>
#include <mutex>
#include <atomic>
#include <unordered_set>

using namespace std;
using namespace Tins;

extern bool debugMode;

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


struct MacMetadata {
    int detectionCount;
    double averageSignalStrenght;
    string signature;
    vector<int> typeCount;
    vector<string> ssidProbes;
    vector<string> htCapabilities;
    string extendedHTCapabilities;
    vector<string> tags;
    vector<string> supportedRates;
};


class PacketManager {
   private:
    map<mac, MacMetadata>* detectedMacs;
    unordered_set<mac>* personalMacs;
    bool disableBackendUpload = false;
    int currentWindowStartTime;
    int secondsPerWindow;
    mutex uploadingMutex;
    bool showPackets;
    string deviceID;
    SQLite::Database *db;

    void uploadToBackend();

    void uploader();

    void syncPersonalMacs();

    void countDevice(mac macAddress, double signalStrength, string ssidProbe,
                                string htCapabilities, string htExtendedCapabilities,
                                vector<int> tags, vector<float> supportedRates, int type);

    void registerManagement(Dot11ManagementFrame *managementFrame, double signalStrength);

    void registerControl(Dot11Control *controlFrame, double signalStrength);

    void registerData(Dot11Data *dataFrame, double signalStrength);

   public:
    PacketManager(bool uploadBackend, string deviceID, bool showPackets, int secondsPerWindow);

    void registerFrame(Packet frame);

};

#endif