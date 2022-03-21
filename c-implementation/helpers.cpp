#include "helpers.h"

#include <curl/curl.h>
#include <tins/tins.h>
#include <json.hpp>

#include <chrono>

#include "main.h"
#include <iostream>

using namespace std::chrono;
using namespace std;
using json = nlohmann::json;

int getCurrentTime() {
    return duration_cast<seconds>(system_clock::now().time_since_epoch())
        .count();
}

mac getStationMAC(Tins::Dot11Data *frame) {
    bool from = frame->from_ds();
    bool to = frame->to_ds();

    if (!to && !from) {
        return frame->addr2();
    } else if (!to && from) {
        return frame->addr1();
    } else if (to && !from) {
        return frame->addr2();
    } else {
        return mac(nullptr);
    }
}

bool isMacValid(mac address) { return true; }

bool isMacFake(mac address) { return (address[0] & 0x02) == 0x02; }

void postJSON(string url, json j) {
    CURL *curl;
    curl = curl_easy_init();

    string jsonString = j.dump();
    if (debugMode) cout << jsonString << endl;

    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, jsonString.c_str());
        curl_easy_setopt(curl, CURLOPT_POST, 1L);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);
        CURLcode res = curl_easy_perform(curl);

        if (res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() failed: %s\n",
                    curl_easy_strerror(res));
        }

        curl_easy_cleanup(curl);
    }
}