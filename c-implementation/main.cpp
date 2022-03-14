#include "main.h"
#include "helpers.h"

#include <curl/curl.h>
#include <stdio.h>
#include <tins/tins.h>

#include <bitset>
#include <iostream>
#include <json.hpp>
#include <set>
#include <string>
#include <vector>
#include <stdlib.h>

using namespace Tins;
using namespace std;
using json = nlohmann::json;

const string HOSTNAME = "tfg-server.raporpe.dev:2000";

void PacketManager::uploadToBackend() {
    CURL *curl;
    curl = curl_easy_init();

    json j;
    j["device_id"] = this->device_id;
    j["count"] = getActiveDevices();

    string jsonString = j.dump();
    string url = "http://" + HOSTNAME + "/v1/ocupation";

    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, jsonString.c_str());
        curl_easy_setopt(curl, CURLOPT_POST, 1L);
        if (uploadBackend) {
            CURLcode res = curl_easy_perform(curl);

            if (res != CURLE_OK) {
                fprintf(stderr, "curl_easy_perform() failed: %s\n",
                        curl_easy_strerror(res));
            }
        }
        curl_easy_cleanup(curl);
    }
}

void PacketManager::checkTimeIncrease() {
    // Check if time should advance
    if (getCurrentTime() - currentSecond > FRAME_TIME) {
        // Delete the inactive macs
        int deleted = 0;
        vector<mac> to_delete;

        for (auto pair : store) {
            if (pair.second.none()) {
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
        for (auto pair : store) {
            store[pair.first] = pair.second << 1;
        }

        // Upload to backend
        uploadToBackend();

        currentSecond = getCurrentTime();
    }
}

void PacketManager::addAndTickMac(mac mac_address) {
    checkTimeIncrease();

    if (store.find(mac_address) != store.end()) {
        // Existing mac address, set last bit to true
        store[mac_address][0] = 1;
    } else
    // If does not exist, insert
    {
        // New mac address, register in memory
        store.insert(make_pair(mac_address, bitset<WINDOW_SIZE>(1)));
    }
}

void PacketManager::tickMac(mac mac_address) {
    checkTimeIncrease();

    // If exists in the store
    if (store.find(mac_address) != store.end()) {
        // Existing mac address, set last bit to true
        store[mac_address][0] = 1;
    }

}

int PacketManager::getActiveDevices() {
    // Count the number of active devices
    int count = 0;
    for (auto pair : this->store) {
        if ((float)pair.second.count() / (float)WINDOW_SIZE >=
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
    SnifferConfiguration config;
    config.set_promisc_mode(true);
    config.set_immediate_mode(true);
    
    Sniffer sniffer(argv[2], config);

    printf("Starting...\n");

    PacketManager *packetManager = new PacketManager(argv[1], argv[3]);

    while (true) {
        // cout << "Getting packet..." << endl;
        Packet pkt = sniffer.next_packet();
        // cout << "Got packet. Processing..." << endl;

        
        
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
