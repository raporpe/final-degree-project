#include "main.h"
#include "helpers.h"
#include "CLI11.hpp"
#include "json.hpp"

#include <stdio.h>
#include <tins/tins.h>

#include <bitset>
#include <iostream>
#include <set>
#include <string>
#include <vector>
#include <stdlib.h>
#include <thread>

using namespace Tins;
using namespace std;
using json = nlohmann::json;

const string HOSTNAME = "tfg-server.raporpe.dev:2000";

void PacketManager::uploadToBackend() {
    json j1;
    j1["device_id"] = this->device_id;
    j1["count"] = getActiveDevices();

    json j2;
    j2["device_id"] = this->device_id;
    j2["seconds_per_window"] = FRAME_TIME;
    j2["number_of_windows"] = WINDOW_SIZE;

    json states;
    for(auto kv : this->store) {
        json state;
        state["mac"] = kv.first.to_string();
        state["state"] = kv.second.state.to_string();
        state["signal_strength"] = kv.second.signal_strength;
        states.push_back(state);
    }

    j2["states"] = states;


    string url1 = "http://" + HOSTNAME + "/v1/ocupation";
    string url2 = "http://" + HOSTNAME + "/v1/state";

    if (uploadBackend) {
        postJSON(url1, j1);
    } 
    postJSON(url2, j2);

}


void PacketManager::checkTimeIncrease() {
    // Check if time should advance
    if (getCurrentTime() - currentSecond > FRAME_TIME) {
        // Delete the inactive macs
        int deleted = 0;
        vector<mac> to_delete;

        for (auto pair : store) {
            if (pair.second.state.none()) {
                to_delete.push_back(pair.first);
                deleted++;
            }
        }

        for (auto del : to_delete) {
            store.erase(del);
        }

        // Print the current state
        for (auto pair : store) {
        //    cout << pair.first << " - " << pair.second << " / "
        //         << (float)pair.second.count() / (float)WINDOW_SIZE << endl;
        }

        int count = getActiveDevices();

        cout << "Deleted " << deleted << " records." << endl;
        cout << "Current state size: " << store.size() << endl;
        cout << "Active devices -> " << count << endl;

        // Advance one bit
        for (auto pair = store.begin(); pair != store.end(); pair++ ) {
            pair->second.state << 1;
        }

        // Upload to backend
        uploadToBackend();

        currentSecond = getCurrentTime();
    }
}

void PacketManager::addAndTickMac(mac mac_address, int signal_strength) {
    checkTimeIncrease();

    if (store.find(mac_address) != store.end()) {
        // Existing mac address, set last bit to true
        store[mac_address].state[0] = 1;

        // Average with the recent signal strength
        cout << store[mac_address].signal_strength << endl; // old
        store[mac_address].signal_strength = (int) store[mac_address].signal_strength * 0.9 + signal_strength * 0.1;
        cout << store[mac_address].signal_strength << endl; // new

    } else
    // If does not exist, insert
    {
        // Create store 
        StoreObject toStore;
        toStore.signal_strength = signal_strength;
        toStore.state = bitset<WINDOW_SIZE>(1);

        // New mac address, register in memory
        store.insert(
            make_pair(mac_address, toStore)
        );
    }
}

void PacketManager::tickMac(mac mac_address, int signal_strength) {
    checkTimeIncrease();

    // If exists in the store
    if (store.find(mac_address) != store.end()) {
        // Existing mac address, set last bit to true
        store[mac_address].state[0] = 1;
    }

}

int PacketManager::getActiveDevices() {
    // Count the number of active devices
    int count = 0;
    for (auto pair : this->store) {
        if ((float)pair.second.state.count() / (float)WINDOW_SIZE >=
            ACTIVITY_PERCENTAGE)
            count++;
    }
    return count;
}


PacketManager::PacketManager(char *upload_backend, char* device_id) {
    string upload(upload_backend);
    this->uploadBackend = upload == "yes";
    this->device_id = device_id;
}

void PacketManager::registerProbeRequest(Dot11ProbeRequest *frame) {
    mac station_address = frame->addr2();
    int signal_strength = frame->bss
    addAndTickMac(station_address);
}

void PacketManager::registerProbeResponse(Dot11ProbeResponse *frame) {
    mac station_address = frame->addr2();
    addAndTickMac(station_address);
}
void PacketManager::registerControl(Dot11Control *frame) {
    mac station_address = frame->addr1();
    tickMac(station_address);
}

void PacketManager::registerData(Dot11Data *frame) {
    mac stationAddress = getStationMAC(frame);
    addAndTickMac(stationAddress);
}

//void PacketManager::registerData(Dot11QoSData *frame) {
//    mac stationAddress = getStationMAC(frame);
//    addAndTickMac(stationAddress);
//}

int main(int argc, char *argv[]) {

    // CLI parsing
    CLI::App app{"C++ data sniffer and storer"};
    string interface = "";
    string device_id = "";
    bool disable_upload = false;
    bool debug_mode = false;

    app.add_option("-i,--interface,--iface", interface, "The 802.11 interface to sniff data from")->required();
    app.add_option("-d,--device,--dev", device_id, "The 802.11 interface to sniff data from")->required();
    app.add_flag("-n,--no-upload", upload, "Disable sending data to backend");
    app.add_flag("--debug", debug_mode, "Enable debug mode");

    CLI11_PARSE(app, argc, argv);

    // Print important information
    cout << "-----------------------" << endl;
    cout << "Capture device: " << interface << endl;
    cout << "Device ID: " << device_id << endl;
    if (debug_mode) cout << "Debug mode enabled!" << endl;
    if (diable_upload) cout << "UPLOAD TO BACKEND DISABLED!" << endl;
    cout << "-----------------------" << endl;

    // Show this message for a second
    thread::sleep_for(seconds(1));

    cout << "Starting capture..." << endl;


    SnifferConfiguration config;
    config.set_promisc_mode(true);
    config.set_immediate_mode(true);
    
    Sniffer sniffer(interface, config);


    PacketManager *packetManager = new PacketManager(disable_upload, );

    while (true) {
        // cout << "Getting packet..." << endl;
        Packet pkt = sniffer.next_packet();
        // cout << "Got packet. Processing..." << endl;

        int signal_strength = pkt.pdu()->find_pdu<RadioTap>()->dbm_signal();
        
        if (Dot11ManagementFrame *p =
                       pkt.pdu()->find_pdu<Dot11ManagementFrame>()) {
            if(p->subtype() != 8 && p->subtype() != 4 && p->subtype() != 5 && p->subtype() != 12) cout << "Management frame -> " << (int)p->subtype() << " mac " << p->addr2() << endl;
        } else if (Dot11ProbeRequest *p = pkt.pdu()->find_pdu<Dot11ProbeRequest>()) {
            // cout << "Probe request -> " << p->addr2() << " with SSID " <<
            // p->ssid() << endl;
            packetManager->registerProbeRequest(p);
        } else if (Dot11ProbeResponse *p =
                       pkt.pdu()->find_pdu<Dot11ProbeResponse>()) {
            // cout << "Probe response -> " << p->addr2() << " with SSID " <<
            // p->ssid() << endl;
            packetManager->registerProbeResponse(p);
        } else if (Dot11Control *p = pkt.pdu()->find_pdu<Dot11Control>()) {
            //cout << "Control frame -> " << p->addr1() << " subtype " << (int) p->subtype() << endl;
            packetManager->registerControl(p);
        //} else if (Dot11QoSData *p = pkt.pdu()->find_pdu<Dot11QoSData>()) {
        //    packetManager->registerData(p);
        //    cout << "Qos" << endl;
        //
        } else if (Dot11Data *p = pkt.pdu()->find_pdu<Dot11Data>()) {
            mac stationAddress = getStationMAC(p);
            //cout << "Data detected with mac " << stationAddress << endl;
            packetManager->registerData(p);
        }
    }
}
