#include "main.h"

#include <stdio.h>
#include <stdlib.h>
#include <tins/tins.h>
#include <unistd.h>

#include <bitset>
#include <iostream>
#include <set>
#include <string>
#include <thread>
#include <vector>

#include "CLI11.hpp"
#include "helpers.h"
#include "json.hpp"

using namespace Tins;
using namespace std;
using json = nlohmann::json;

const string HOSTNAME = "https://tfg-api.raporpe.dev";
bool debugMode;

void PacketManager::uploadToBackend() {
    json j;
    j["device_id"] = this->deviceID;
    j["seconds_per_window"] = WINDOW_TIME;
    j["number_of_windows"] = RECORD_SIZE;
    j["time"] = this->currentStateStartTime;

    json macs;
    for (auto kv : *detectedMacs) {
        json macMetadata;
        macMetadata["detection_count"] = kv.second.detectionCount;
        macMetadata["average_signal_strength"] =
            kv.second.averageSignalStrenght;
        macMetadata["signature"] = kv.second.signature;
        macMetadata["type_count"] = kv.second.typeCount;
        macs[kv.first.to_string()] = macMetadata;
    }

    j["detected_macs"] = macs;

    string url = HOSTNAME + "/v1/state";

    if (!disableBackendUpload) postJSON(url, j);
}

void PacketManager::uploader() {
    while (true) {
        // Check if time should advance
        if (getCurrentTime() % WINDOW_TIME == 0) {
            // Lock the mutex to avoid modifying the store
            if (debugMode) cout << "Time! adquiring mutex..." << endl;
            this->uploadingMutex.lock();
            if (debugMode) cout << "Mutex adquired!" << endl;

            cout << "-----------------------------------" << endl;
            cout << "Current pointed macs: " << personalDeviceMacs->size()
                 << endl;
            cout << "Temporary saved macs: " << detectedMacs->size() << endl;
            cout << "-----------------------------------" << endl;

            // Upload to backend
            uploadToBackend();

            // Clear the current detectedMacs
            delete detectedMacs;
            detectedMacs = new map<mac, MacMetadata>();

            currentStateStartTime += WINDOW_TIME;

            this->uploadingMutex.unlock();
            if (debugMode) cout << "Mutex released" << endl;
        }
        this_thread::sleep_for(chrono::seconds(1));
    }
}

void PacketManager::countDevice(mac macAddress, int signalStrength, int type) {
    // Do not allow invalid macs (multicast and broadcast)
    if (!isMacValid(macAddress)) return;

    // Check for invalid type
    if (type < 0 || type > 2) return;

    this->uploadingMutex.lock();

    bool macInPersonalDevice =
        personalDeviceMacs->find(macAddress) != personalDeviceMacs->end();

    if (!macInPersonalDevice) {
        if (type == Dot11::CONTROL) {
            // We cannot make sure if the mac is from a personal device
            // or if it is from a bssid
            this->uploadingMutex.unlock();
            return;
        } else {
            // Register the mac in the personal devices list
            personalDeviceMacs->insert(macAddress);
        }
    }

    // If the mac address has already been counted in this window
    if (detectedMacs->find(macAddress) != detectedMacs->end()) {
        // Get the current signal average and detection count
        int oldAverageSignal =
            detectedMacs->find(macAddress)->second.averageSignalStrenght;
        int detectionCount =
            detectedMacs->find(macAddress)->second.detectionCount;

        detectedMacs->find(macAddress)->second.averageSignalStrenght =
            oldAverageSignal * (signalStrength - oldAverageSignal) /
            detectionCount;

        // Increase the detection count
        detectedMacs->find(macAddress)->second.detectionCount++;

        // Increase the count of the type
        detectedMacs->find(macAddress)->second.typeCount[type]++;

        // If the mac address has not been counted in the current window
    } else {
        MacMetadata macMetadata;
        macMetadata.averageSignalStrenght = signalStrength;
        macMetadata.detectionCount = 1;
        macMetadata.signature = "";
        macMetadata.typeCount = vector<int>(3, 0);
        macMetadata.typeCount[type] = 1;

        // Insert in the detected macs
        detectedMacs->insert(make_pair(macAddress, macMetadata));
    }

    this->uploadingMutex.unlock();
}

PacketManager::PacketManager(bool uploadBackend, string deviceID,
                             bool showPackets) {
    this->disableBackendUpload = uploadBackend;
    this->deviceID = deviceID;
    this->currentStateStartTime =
        getCurrentTime() - (getCurrentTime() % WINDOW_TIME);
    this->personalDeviceMacs = new unordered_set<mac>();
    this->detectedMacs = new map<mac, MacMetadata>();
    this->showPackets = showPackets;

    // Start the uploader thread
    thread upload(&PacketManager::uploader, this);
    upload.detach();
}

void PacketManager::registerFrame(Packet frame) {
    int signalStrength = frame.pdu()->find_pdu<RadioTap>()->dbm_signal();
    Dot11 *dot11Frame = frame.pdu()->find_pdu<Dot11>();
    cout << (int)dot11Frame->type() << endl;
    switch (dot11Frame->type()) {
        case Dot11::MANAGEMENT:
            registerManagement(dot11Frame->find_pdu<Dot11ManagementFrame>(),
                               signalStrength);
            break;
        case Dot11::CONTROL:
            registerControl(dot11Frame->find_pdu<Dot11Control>(),
                            signalStrength);
            break;
        case Dot11::DATA:
            registerData(dot11Frame->find_pdu<Dot11Data>(), signalStrength);
            break;
    }
}

void PacketManager::registerManagement(Dot11ManagementFrame *managementFrame,
                                       int signalStrength) {
    if (managementFrame->subtype() == Dot11::ManagementSubtypes::PROBE_REQ) {
        mac stationAddress = managementFrame->addr2();
        if (showPackets) {
            cout << "Probe request  -> " << managementFrame->addr2()
                 << " with SSID " << managementFrame->ssid() << endl;
        }
        countDevice(stationAddress, signalStrength, Dot11::MANAGEMENT);

    } else if (managementFrame->subtype() ==
               Dot11::ManagementSubtypes::PROBE_RESP) {
        mac stationAddress = managementFrame->addr2();

        if (showPackets) {
            cout << "Probe response -> " << managementFrame->addr2()
                 << " with SSID " << managementFrame->ssid() << endl;
        }
        countDevice(stationAddress, signalStrength, Dot11::MANAGEMENT);

    } else if (debugMode && managementFrame->subtype() != 8 &&
               managementFrame->subtype() != 4 &&
               managementFrame->subtype() != 5 &&
               managementFrame->subtype() != 12) {
        cout << "!Mngmnt frame  -> mac " << managementFrame->addr2()
             << " subtype " << (int)managementFrame->subtype() << endl;
    }
}

void PacketManager::registerControl(Dot11Control *controlFrame,
                                    int signalStrength) {
    if (true) {
        cout << "Control frame  -> " << controlFrame->addr1() << " subtype "
             << (int)controlFrame->subtype() << endl;
    }

    mac address = controlFrame->addr1();
    countDevice(address, signalStrength, Dot11::CONTROL);
}

void PacketManager::registerData(Dot11Data *dataFrame, int signalStrength) {
    if (showPackets) {
        mac stationAddress = getStationMAC(dataFrame);
        cout << "Data frame     -> " << stationAddress << " subtype "
             << (int)dataFrame->subtype() << endl;
    }

    mac stationAddress = getStationMAC(dataFrame);
    countDevice(stationAddress, signalStrength, Dot11::DATA);
}

int main(int argc, char *argv[]) {
    // CLI parsing
    CLI::App app{"C++ data sniffer and storer"};
    string interface = "";
    string deviceID = "";
    bool disableUpload = false;
    bool showPackets = false;

    app.add_option("-i,--interface,--iface", interface,
                   "The 802.11 interface to sniff data from")
        ->required();
    app.add_option("-d,--device,--dev", deviceID,
                   "The 802.11 interface to sniff data from")
        ->required();
    app.add_flag("-n,--no-upload", disableUpload,
                 "Disable sending data to backend");
    app.add_flag("--debug", debugMode, "Enable debug mode");
    app.add_flag("-p,--packets", showPackets, "Show all the captured packets");

    CLI11_PARSE(app, argc, argv);

    bool sudo = geteuid() == 0;
    if (!sudo) {
        cout << "You must run this program as root!" << endl;
    }

    // Get config from backend
    json backendConfig = getJSON(HOSTNAME + "/v1/config");
    int w_size = backendConfig["window_size"];
    int w_time = backendConfig["window_time"];

    // Print important information
    cout << "-----------------------" << endl;
    cout << "Capture device: " << interface << endl;
    cout << "Device ID: " << deviceID << endl;
    if (debugMode) cout << "Debug mode enabled!" << endl;
    if (disableUpload) cout << "UPLOAD TO BACKEND DISABLED!" << endl;
    cout << "Window size: " << w_size << endl;
    cout << "Window time: " << w_time << endl;
    cout << "-----------------------" << endl;

    cout << "Enabling monitor mode in interface " << interface << "..." << endl;
    if (!is_monitor_mode(interface)) {
        // Try to set in monitor mode
        set_monitor_mode(interface);
        if (!is_monitor_mode(interface)) {
            cout << "Could not enable monitor mode in interface "
                 << interface << endl;
            exit(1);
        }
    }

    cout << "Starting channel switcher..." << endl;

    thread t1(channel_switcher, interface);
    t1.detach();

    cout << "Starting capture..." << endl;

    // Show previous messages for three seconds
    this_thread::sleep_for(chrono::seconds(3));

    // Actually start the sniffer
    SnifferConfiguration config;
    config.set_promisc_mode(true);
    config.set_immediate_mode(true);

    Sniffer sniffer(interface, config);

    PacketManager *packetManager =
        new PacketManager(disableUpload, deviceID, showPackets);

    while (true) {
        Packet pkt = sniffer.next_packet();
        packetManager->registerFrame(pkt);
    }
}
