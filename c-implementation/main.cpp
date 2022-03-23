#include "main.h"
#include "helpers.h"
#include "CLI11.hpp"
#include "json.hpp"

#include <stdio.h>
#include <tins/tins.h>
#include <unistd.h>

#include <bitset>
#include <iostream>
#include <set>
#include <string>
#include <vector>
#include <stdlib.h>
#include <thread>
#include <regex>

using namespace Tins;
using namespace std;
using json = nlohmann::json;

const string HOSTNAME = "tfg-server.raporpe.dev:2000";
bool debugMode;

void PacketManager::uploadToBackend() {

    json j;
    j["device_id"] = this->deviceID;
    j["seconds_per_window"] = WINDOW_TIME;
    j["number_of_windows"] = RECORD_SIZE;
    j["time"] = this->currentStateStartTime;

    json states;
    for(auto kv : this->store) {
        json state;
        state["mac"] = kv.first.to_string();
        state["record"] = kv.second.record.to_string();
        state["signal_strength"] = kv.second.signalStrength;
        states.push_back(state);
    }

    j["mac_states"] = states;

    string url = "http://" + HOSTNAME + "/v1/state";

    if (!disableBackendUpload) postJSON(url, j);

}


void PacketManager::uploader() {
    while(true) {
        // Check if time should advance
        if (getCurrentTime() % WINDOW_TIME == 0) {
            // Lock the mutex to avoid modifying the store
            if (debugMode) cout << "Time! adquiring mutex..." << endl;
            this->uploadingMutex.lock();
            if (debugMode) cout << "Mutex adquired!" << endl;
            
            // Delete the inactive macs
            int deleted = 0;
            vector<mac> to_delete;

            for (auto pair : store) {
                if (pair.second.record.none()) {
                    to_delete.push_back(pair.first);
                    deleted++;
                }
            }

            for (auto del : to_delete) {
                store.erase(del);
            }

            int count = getActiveDevices();
            cout << "-----------------------------------" << endl;
            cout << "Current state size: " << store.size() << endl;
            cout << "Active devices -> " << count << endl;
            cout << "Deleted " << deleted << " records." << endl;
            cout << "-----------------------------------" << endl;

            // Advance one bit
            for (auto& pair : store) {
                pair.second.record <<= 1;
            }

            // Upload to backend
            uploadToBackend();

            currentStateStartTime += WINDOW_TIME;

            this->uploadingMutex.unlock();
            if (debugMode) cout << "Mutex released" << endl;
        }
        this_thread::sleep_for(chrono::seconds(1));
    }
}

void PacketManager::addAndTickMac(mac macAddress, int signalStrength) {
    // Do not allow invalid macs (multicast and broadcast)
    if (!isMacValid(macAddress)) return;

    this->uploadingMutex.lock();

    if (store.find(macAddress) != store.end()) {
        // Existing mac address, set last bit to true
        store[macAddress].record[0] = 1;

        // Average with the recent signal strength
        store[macAddress].signalStrength = (int) store[macAddress].signalStrength * 0.9 + signalStrength * 0.1;

    } else
    // If does not exist, insert
    {
        // Create store 
        RecordObject toStore;
        toStore.signalStrength = signalStrength;
        toStore.record = bitset<RECORD_SIZE>(1);

        // New mac address, register in memory
        store.insert(
            make_pair(macAddress, toStore)
        );
    }
    this->uploadingMutex.unlock();

}

void PacketManager::tickMac(mac macAddress, int signalStrength) {
    this->uploadingMutex.lock();

    // If exists in the store
    if (store.find(macAddress) != store.end()) {
        // Existing mac address, set last bit to true
        store[macAddress].record[0] = 1;
        store[macAddress].signalStrength = (int) store[macAddress].signalStrength * 0.9 + signalStrength * 0.1;
    }

    this->uploadingMutex.unlock();
}

int PacketManager::getActiveDevices() {
    // Count the number of active devices
    int count = 0;
    for (auto pair : this->store) {
        if ((float)pair.second.record.count() / (float)RECORD_SIZE >=
            ACTIVITY_PERCENTAGE)
            count++;
    }
    return count;
}


PacketManager::PacketManager(bool uploadBackend, string deviceID) {
    this->disableBackendUpload = uploadBackend;
    this->deviceID = deviceID;
    this->currentStateStartTime = getCurrentTime() - (getCurrentTime() % WINDOW_TIME);

    // Start the uploader thread
    thread upload(&PacketManager::uploader, this);
    upload.detach();
}

void PacketManager::registerProbeRequest(Dot11ProbeRequest *frame, int signalStrength) {
    mac stationAddress = frame->addr2();
    addAndTickMac(stationAddress, signalStrength);
}

void PacketManager::registerProbeResponse(Dot11ProbeResponse *frame, int signalStrength) {
    mac stationAddress = frame->addr2();
    addAndTickMac(stationAddress, signalStrength);
}

void PacketManager::registerControl(Dot11Control *frame, int signalStrength) {
    mac stationAddress = frame->addr1();
    tickMac(stationAddress, signalStrength);
}

void PacketManager::registerData(Dot11Data *frame, int signalStrength) {
    mac stationAddress = getStationMAC(frame);
    addAndTickMac(stationAddress, signalStrength);
}

void channel_switcher(string interface) {
    const vector<int> channels = {1,6,11,2,7,12,3,9,13,4,10,5,8};

    // Switch channels for ever
    while(true) {
        for (auto channel : channels) {
            string command = "iw dev " + interface + " set channel " + to_string(channel);
            system(command.c_str());
            this_thread::sleep_for(chrono::milliseconds(100));
        }
    }
}

void set_monitor_mode(string interface) {
    string interface_down = "ip link set " + interface + " down";
    string interface_up = "ip link set " + interface + " up";
    string set_monitor = "iw " + interface + " set monitor control";
    system(interface_down.c_str());
    system(set_monitor.c_str());
    system(interface_up.c_str());
}

bool is_monitor_mode(string interface) {
    // Command that calls iw to get the interface
    string cmd = "iw dev " + interface + " info";

    array<char, 128> buffer;
    string result;
    unique_ptr<FILE, decltype(&pclose)> pipe(popen(cmd.c_str(), "r"), pclose);
    if (!pipe) {
        throw runtime_error("popen() failed!");
    }
    while (fgets(buffer.data(), buffer.size(), pipe.get()) != nullptr) {
        result += buffer.data();
    }

    // Match with regex
    smatch match;
    regex rx("(managed|monitor)");
    regex_search(result, match, rx);
    string res(match[0]);

    return res == "monitor";

}

int main(int argc, char *argv[]) {

    // CLI parsing
    CLI::App app{"C++ data sniffer and storer"};
    string interface = "";
    string deviceID = "";
    bool disableUpload = false;
    bool showPackets = false;
    bool sudo = geteuid() == 0;

    app.add_option("-i,--interface,--iface", interface, "The 802.11 interface to sniff data from")->required();
    app.add_option("-d,--device,--dev", deviceID, "The 802.11 interface to sniff data from")->required();
    app.add_flag("-n,--no-upload", disableUpload, "Disable sending data to backend");
    app.add_flag("--debug", debugMode, "Enable debug mode");
    app.add_flag("-p,--packets", showPackets, "Show all the captured packets");

    CLI11_PARSE(app, argc, argv);

    if(!sudo) {
        cout << "You must run this program as root!" << endl;
    }

    // Get config from backend
    json backendConfig = getJSON("http://" + HOSTNAME + "/v1/config");
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
        if(!is_monitor_mode(interface)) {
            cout << "Could not enable monitor mode in interface " << interface << endl;
            exit(1);
        } 
    }


    // Show this message for a second
    this_thread::sleep_for(chrono::seconds(1));

    cout << "Starting channel switcher..." << endl;

    thread t1(channel_switcher, interface);
    t1.detach();

    cout << "Starting capture..." << endl;

    SnifferConfiguration config;
    config.set_promisc_mode(true);
    config.set_immediate_mode(true);
    
    Sniffer sniffer(interface, config);

    PacketManager *packetManager = new PacketManager(disableUpload, deviceID);

    while (true) {
        Packet pkt = sniffer.next_packet();
        int signalStrength = pkt.pdu()->find_pdu<RadioTap>()->dbm_signal();
        
        if (Dot11ManagementFrame *p =
                       pkt.pdu()->find_pdu<Dot11ManagementFrame>()) {
            if(debugMode && p->subtype() != 8 && p->subtype() != 4 && p->subtype() != 5 && p->subtype() != 12) cout << "Management frame -> " << (int)p->subtype() << " mac " << p->addr2() << endl;
        } else if (Dot11ProbeRequest *p = pkt.pdu()->find_pdu<Dot11ProbeRequest>()) {
            if (showPackets) cout << "Probe request  -> " << p->addr2() << " with SSID " << p->ssid() << endl;
            packetManager->registerProbeRequest(p, signalStrength);
        } else if (Dot11ProbeResponse *p =
                       pkt.pdu()->find_pdu<Dot11ProbeResponse>()) {
            if (showPackets) cout << "Probe response -> " << p->addr2() << " with SSID " << p->ssid() << endl;
            packetManager->registerProbeResponse(p, signalStrength);
        } else if (Dot11Control *p = pkt.pdu()->find_pdu<Dot11Control>()) {
            if (showPackets) cout << "Control frame  -> " << p->addr1() << " subtype " << (int) p->subtype() << endl;
            packetManager->registerControl(p, signalStrength);
        } else if (Dot11Data *p = pkt.pdu()->find_pdu<Dot11Data>()) {
            mac stationAddress = getStationMAC(p);
            if (showPackets) cout << "Data frame     -> " << stationAddress << " subtype " << (int) p->subtype() << endl;
            packetManager->registerData(p, signalStrength);
        }
    }
}
