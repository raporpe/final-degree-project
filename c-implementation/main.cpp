#include <curl/curl.h>
#include <stdio.h>
#include <tins/tins.h>

#include <bitset>
#include <ctime>
#include <iostream>
#include <json.hpp>
#include <set>
#include <string>
#include <vector>

using namespace Tins;
using namespace std;
using namespace chrono;
using json = nlohmann::json;

typedef HWAddress<6> mac;

const int WINDOW_SIZE = 15;
const int FRAME_TIME = 60;
const float ACTIVITY_PERCENTAGE = 0.6;

struct ProbeRequest {
    string stationMac;
    string intent;
    string time;
    int frequency;
    int power;
};

struct UploadJSONData {
    string deviceID;
    vector<ProbeRequest> probeRequests;
};

class PacketManager {
   private:
    vector<Dot11ProbeResponse> probeResponses;
    vector<Dot11ProbeRequest> probeRequests;
    int currentSecond = 0;
    map<mac, bitset<WINDOW_SIZE>> store;
    bool uploadBackend = false;

    int getCurrentTime() {
        return duration_cast<seconds>(system_clock::now().time_since_epoch())
            .count();
    }

    void uploadToBackend() {
        CURL *curl;
        curl = curl_easy_init();

        json j;
        j["device_id"] = "raspberry-1";
        j["count"] = getActiveDevices();
        cout << j.dump() << endl;

        string jsonString = j.dump();
        cout << jsonString << endl;

        if (curl) {
            curl_easy_setopt(curl, CURLOPT_URL,
                             "http://tfg-server.raporpe.dev:2000/v1/ocupation");
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

    void checkTimeIncrease() {
        // Check if time should advance
        if (this->getCurrentTime() - currentSecond > FRAME_TIME) {
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
                cout << pair.first << " - " << pair.second << " / "
                     << (float)pair.second.count() / (float)WINDOW_SIZE << endl;
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

            currentSecond = this->getCurrentTime();
        }
    }

    void addAndTickMac(mac mac_address) {
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

    int getActiveDevices() {
        // Count the number of active devices
        int count = 0;
        for (auto pair : this->store) {
            if ((float)pair.second.count() / (float)WINDOW_SIZE >=
                ACTIVITY_PERCENTAGE)
                count++;
        }
        return count;
    }

    void tickMac(mac mac_address) {
        checkTimeIncrease();

        // If exists in the store
        if (store.find(mac_address) != store.end()) {
            // Existing mac address, set last bit to true
            store[mac_address][0] = 1;
        }

        // If does not exist, do nothing
    }

   public:
    PacketManager(char *upload_backend) {
        string upload(upload_backend);
        this->uploadBackend = upload == "yes";
    }

    void registerProbeRequest(Dot11ProbeRequest *frame) {
        mac station_address = frame->addr2();
        addAndTickMac(station_address);
    }

    void registerProbeResponse(Dot11ProbeResponse *frame) {
        mac station_address = frame->addr2();
        addAndTickMac(station_address);
    }
    void registerControl(Dot11Control *frame) {
        mac station_address = frame->addr1();
        tickMac(station_address);
    }
};

int main(int argc, char *argv[]) {
    SnifferConfiguration config;
    config.set_promisc_mode(true);
    config.set_immediate_mode(true);

    Sniffer sniffer("wlan1", config);

    printf("Starting...\n");

    PacketManager *packetManager = new PacketManager(argv[1]);

    while (true) {
        // cout << "Getting packet..." << endl;
        Packet pkt = sniffer.next_packet();
        // cout << "Got packet. Processing..." << endl;

        if (Dot11ProbeRequest *p = pkt.pdu()->find_pdu<Dot11ProbeRequest>()) {
            // cout << "Probe request -> " << p->addr2() << " with SSID " <<
            // p->ssid() << endl;
            packetManager->registerProbeRequest(p);
        } else if (Dot11ProbeResponse *p =
            pkt.pdu()->find_pdu<Dot11ProbeResponse>()) {
            // cout << "Probe response -> " << p->addr2() << " with SSID " <<
            // p->ssid() << endl;
            packetManager->registerProbeResponse(p);
        } else if (Dot11Control *p = pkt.pdu()->find_pdu<Dot11Control>()) {
            // cout << "Control frame -> " << p->addr1() << endl;
            packetManager->registerControl(p);
        } else if (Dot11ControlTA *p = pkt.pdu()->find_pdu<Dot11ControlTA>()) {
            cout << "Control detected" << endl;
        }
    }
}
