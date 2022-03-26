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

bool isMacValid(mac address) {
    bool isNull = address == mac(nullptr);
    return address.is_unicast() && !isNull; 
}

bool isMacFake(mac address) { return (address[0] & 0x02) == 0x02; }

json postJSON(string url, json j) {
    auto curl = curl_easy_init();

    string jsonString = j.dump();
    if (debugMode) cout << jsonString << endl;

    string response;

    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, jsonString.c_str());
        curl_easy_setopt(curl, CURLOPT_POST, 1L);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, curlWriteCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);
        CURLcode res = curl_easy_perform(curl);

        if (res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() failed: %s\n",
                    curl_easy_strerror(res));
        }

        curl_easy_cleanup(curl);
    }

    return json::parse(response);
}

size_t curlWriteCallback(void *contents, size_t size, size_t nmemb, std::string *s)
{
    size_t newLength = size*nmemb;
    s->append((char*)contents, newLength);

    return newLength;
}

json getJSON(string url) {
    auto curl = curl_easy_init();
    string response;
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, curlWriteCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);

        CURLcode res = curl_easy_perform(curl);
        if (res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() failed: %s\n",
                    curl_easy_strerror(res));
        }
        curl_easy_cleanup(curl);

    }
    return json::parse(response);
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

