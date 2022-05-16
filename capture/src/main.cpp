#include "main.h"

#include <stdio.h>
#include <stdlib.h>
#include <tins/tins.h>
#include <unistd.h>

#include <climits>

#include <bitset>
#include <ctime>
#include <iostream>
#include <set>
#include <string>
#include <thread>
#include <vector>
#include <fmt/format.h>

#include "helpers.h"
#include "CLI11.hpp"
#include "json.hpp"

using namespace Tins;
using namespace std;
using json = nlohmann::json;

const string HOSTNAME = "http://tfg-api.raporpe.dev:8080";
bool debugMode;

void PacketManager::uploadToBackend()
{
    json j;
    j["device_id"] = this->deviceID;
    j["seconds_per_window"] = this->secondsPerWindow;
    j["start_time"] = this->currentWindowStartTime;
    j["end_time"] = this->currentWindowStartTime + this->secondsPerWindow;

    json macs;
    for (auto kv : *detectedMacs)
    {
        json macMetadata;
        macMetadata["detection_count"] = kv.second.detectionCount;
        macMetadata["average_signal_strength"] =
            kv.second.averageSignalStrenght;
        macMetadata["type_count"] = kv.second.typeCount;
        macMetadata["ssid_probes"] = kv.second.ssidProbes;

        if (kv.second.htCapabilities == "") {
            macMetadata["ht_capabilities"] = nullptr;
        } else {
            macMetadata["ht_capabilities"] = kv.second.htCapabilities;
        }

        if (kv.second.htExtendedCapabilities == "") {
            macMetadata["ht_extended_capabilities"] = nullptr;
        } else {
            macMetadata["ht_extended_capabilities"] = kv.second.htExtendedCapabilities;
        }

        macMetadata["supported_rates"] = kv.second.supportedRates;
        macMetadata["tags"] = kv.second.tags;
        macs[kv.first.to_string()] = macMetadata;
    }

    j["detected_macs"] = macs;

    string url = HOSTNAME + "/v1/detected-macs";

    if (!disableBackendUpload)
    {
        try
        {
            postJSON(url, j);
        }
        catch (UnavailableBackendException &e)
        {
            // Save the json in sqlite for sending it later
            cout << "Inserting json in DB!" << endl;
            SQLite::Statement query(*this->db, "INSERT INTO WINDOWS (json) VALUES ( ? );");
            query.bind(1, j.dump());
            while (query.executeStep())
            {
                cout << "step" << endl;
            }
        }
    }

    // Try to upload old jsons stored in the databse

    SQLite::Statement query(*this->db, "SELECT * FROM WINDOWS");
    while (query.executeStep())
    {
        int id = query.getColumn(0);
        string storedJSON = query.getColumn(1).getText();
        // Try to send to backend
        bool correctPost = true;
        cout << "aaaaaaaaaaaaa" << endl
             << storedJSON << endl
             << "aaaaaaaaaa" << endl;
        try
        {
            postJSON(url, json::parse(storedJSON));
        }
        catch (UnavailableBackendException &e)
        {
            correctPost = false;
            break;
        }

        if (correctPost)
        {
            this->db->exec("DELETE FROM WINDOWS WHERE ID = '" + to_string(id) + "'");
        }

        cout << "Trying to restore json with id " << id << endl;
    }
}

void PacketManager::syncPersonalMacs()
{
    // Send the current macs first
    json j;
    j["device_id"] = this->deviceID;
    json p = json::array();
    for (auto k : *personalMacs)
    {
        p.push_back(k.to_string());
    }
    j["personal_macs"] = p;

    string url = HOSTNAME + "/v1/personal-macs";
    json response = json::array();
    try
    {
        json response = postJSON(url, j);
    }
    catch (UnavailableBackendException &e)
    {
        cout << "Cannot connect with backend. Skipping personal macs sync!"
             << endl;
    }

    // Fill the current macs with the received data
    for (auto mac : response)
    {
        personalMacs->insert(mac.get<string>());
    }
}

void PacketManager::uploader()
{
    while (true)
    {
        // Check if time should advance
        if (getCurrentTime() > currentWindowStartTime + secondsPerWindow)
        {
            // Lock the mutex to avoid modifying the store
            this->uploadingMutex.lock();

            cout << "-----------------------------------" << endl;
            cout << "Personal devices index size: " << personalMacs->size()
                 << endl;
            cout << "Detected macs for current window: " << detectedMacs->size()
                 << endl;
            cout << "-----------------------------------" << endl;

            // Upload to backend
            uploadToBackend();

            // sync the personal macs
            syncPersonalMacs();

            // Clear the current detectedMacs
            delete detectedMacs;
            detectedMacs = new map<mac, MacMetadata>();

            currentWindowStartTime += secondsPerWindow;

            this->uploadingMutex.unlock();
        }
        int sleepFor =
            currentWindowStartTime + secondsPerWindow - getCurrentTime();
        this_thread::sleep_for(chrono::seconds(sleepFor));
    }
}

void PacketManager::countDevice(mac macAddress, double signalStrength, string ssidProbe,
                                string htCapabilities, string htExtendedCapabilities,
                                vector<int> tags, vector<float> supportedRates, int type)
{
    // Do not allow invalid macs (multicast and broadcast)
    if (!isMacValid(macAddress))
        return;

    // Check for invalid type
    if (type < 0 || type > 2)
        return;

    this->uploadingMutex.lock();

    bool macInPersonalDevice =
        personalMacs->find(macAddress) != personalMacs->end();

    if (!macInPersonalDevice)
    {
        if (type == Dot11::CONTROL)
        {
            // We cannot make sure if the mac is from a personal device
            // or if it is from a bssid
            this->uploadingMutex.unlock();
            return;
        }
        else
        {
            // Register the mac in the personal devices list
            personalMacs->insert(macAddress);
        }
    }

    // If the mac address has already been counted in this window
    if (detectedMacs->find(macAddress) != detectedMacs->end())
    {
        // Get the current signal average and detection count
        double oldAverageSignal =
            detectedMacs->find(macAddress)->second.averageSignalStrenght;
        int detectionCount =
            detectedMacs->find(macAddress)->second.detectionCount;

        detectedMacs->find(macAddress)->second.averageSignalStrenght =
            ((oldAverageSignal * detectionCount) + signalStrength) /
            (detectionCount + 1);

        // Increase the detection count
        detectedMacs->find(macAddress)->second.detectionCount++;

        // Increase the count of the type
        detectedMacs->find(macAddress)->second.typeCount[type]++;

        // Add the ssid probe if it is set
        if (ssidProbe != "")
            detectedMacs->find(macAddress)->second.ssidProbes.push_back(ssidProbe);
        
        // Add htCapabilities
        if (htCapabilities != "")
            detectedMacs->find(macAddress)->second.htCapabilities = htCapabilities;

        // HT Extended capabilities
        if (htExtendedCapabilities != "")
            detectedMacs->find(macAddress)->second.htExtendedCapabilities = htExtendedCapabilities;

        // Supported rates
        copy(supportedRates.begin(), supportedRates.end(), back_inserter(detectedMacs->find(macAddress)->second.supportedRates));

        // Tags
        copy(tags.begin(), tags.end(), back_inserter(detectedMacs->find(macAddress)->second.tags));

        // If the mac address has not been counted in the current window
    }
    else
    {
        MacMetadata macMetadata;
        macMetadata.averageSignalStrenght = signalStrength;
        macMetadata.detectionCount = 1;
        macMetadata.typeCount = vector<int>(3, 0);
        macMetadata.typeCount[type] = 1;

        // SSID probes
        macMetadata.ssidProbes = vector<string>();
        if (ssidProbe != "")
            macMetadata.ssidProbes.push_back(ssidProbe);

        // HT Capabilities
        macMetadata.htCapabilities = "";
        if (htCapabilities != "")
            macMetadata.htCapabilities = htCapabilities;
        
        // HT Extended capabilities
        macMetadata.htExtendedCapabilities = "";
        if (htExtendedCapabilities != "")
            macMetadata.htExtendedCapabilities = htExtendedCapabilities;

        // Supported rates
        macMetadata.supportedRates = vector<float>();
        if (!supportedRates.empty()) copy(supportedRates.begin(), supportedRates.end(), back_inserter(macMetadata.supportedRates));

        // Tags
        macMetadata.tags = vector<int>();
        if (!tags.empty()) copy(tags.begin(), tags.end(), back_inserter(macMetadata.tags));

        // Insert in the detected macs
        detectedMacs->insert(make_pair(macAddress, macMetadata));
    }

    this->uploadingMutex.unlock();
}

PacketManager::PacketManager(bool uploadBackend, string deviceID,
                             bool showPackets, int secondsPerWindow)
{
    this->disableBackendUpload = uploadBackend;
    this->deviceID = deviceID;
    this->secondsPerWindow = secondsPerWindow;
    this->currentWindowStartTime =
        getCurrentTime() - (getCurrentTime() % secondsPerWindow);
    this->personalMacs = new unordered_set<mac>();
    this->detectedMacs = new map<mac, MacMetadata>();
    this->showPackets = showPackets;

    string homeDir = string(getenv("HOME"));
    mkdir((homeDir + "/tfg_db").c_str(), 0);
    this->db = new SQLite::Database(homeDir + "/tfg_db/main.db", SQLite::OPEN_READWRITE | SQLite::OPEN_CREATE);
    this->db->exec("CREATE TABLE IF NOT EXISTS WINDOWS (id INTEGER PRIMARY KEY AUTOINCREMENT, json TEXT NOT NULL);");

    // Sync the macs with the backend
    this->syncPersonalMacs();

    // Start the uploader thread
    thread upload(&PacketManager::uploader, this);
    upload.detach();
}

void PacketManager::registerFrame(Packet frame)
{
    if (!frame.pdu()->find_pdu<RadioTap>())
    {
        return;
    }

    if (!frame.pdu()->find_pdu<Dot11>())
    {
        return;
    }

    double signalStrength = frame.pdu()->find_pdu<RadioTap>()->dbm_signal();
    signalStrength = 1000000000.0 *
                     pow(10, signalStrength / 10.0); // Convert to pW (10^-12)

    if (auto dot11Frame = frame.pdu()->find_pdu<Dot11>())
    {
        switch (dot11Frame->type())
        {
        case Dot11::MANAGEMENT:
            if (auto f = dot11Frame->find_pdu<Dot11ManagementFrame>())
            {
                registerManagement(f, signalStrength);
            }
            break;
        case Dot11::CONTROL:
            if (auto f = dot11Frame->find_pdu<Dot11Control>())
            {
                registerControl(f, signalStrength);
            }
            break;
        case Dot11::DATA:
            if (auto f = dot11Frame->find_pdu<Dot11Data>())
            {
                registerData(f, signalStrength);
            }
            break;
        }
    }
}

vector<int> getOptionsInt(Dot11::options_type opt) {
    vector<int> ret;
    for (Dot11::option o : opt) {
        ret.push_back((int)o.option());
    }
    return ret;
}

void PacketManager::registerManagement(Dot11ManagementFrame *managementFrame,
                                       double signalStrength)
{
    if (managementFrame == nullptr)
    {
        cout << "NULL managementframe!!!" << endl;
    }
    if (managementFrame->subtype() == Dot11::ManagementSubtypes::PROBE_REQ)
    {

        mac stationAddress = managementFrame->addr2();

        // SSID probe inside probe request
        string ssidProbe = managementFrame->ssid();

        // Get the HT Capabilities
        const Dot11::option *htCapabilites = managementFrame->search_option(Dot11::HT_CAPABILITY);
        string ht = "";
        if (htCapabilites)
            ht = fmt::format("{:#x}", (int)*htCapabilites->data_ptr());

        // Get the HT Extended Capabilities
        const Dot11::option *htExtendedCapabilites = managementFrame->search_option(static_cast<Dot11::OptionTypes>(127));
        string htExtended = "";
        if (htExtendedCapabilites)
            htExtended = fmt::format("{:#x}", (int)*htExtendedCapabilites->data_ptr());

        // Get the supported rates
        vector<float> supportedRates = managementFrame->supported_rates();
        vector<float> supportedExtendedRates = managementFrame->supported_rates();
        
        // Copy supportedExtendedRates into supportedRates
        copy(supportedRates.begin(), supportedRates.end(), back_inserter(supportedExtendedRates));
        
        // Get the tags in order
        vector<int> tags = getOptionsInt(managementFrame->options());

        if (showPackets)
        {
            cout << "Probe request  -> " << managementFrame->addr2()
                 << " with SSID " << managementFrame->ssid() << endl;

            const Dot11::option *opt = managementFrame->search_option(Dot11::HT_CAPABILITY);
        }
        countDevice(stationAddress, signalStrength, ssidProbe, ht, htExtended, tags, supportedRates, Dot11::MANAGEMENT);
    }
    else if (managementFrame->subtype() ==
             Dot11::ManagementSubtypes::PROBE_RESP)
    {

        mac stationAddress = managementFrame->addr2();
        string ssidProbe = managementFrame->ssid();

        if (showPackets)
        {
            cout << "Probe response -> " << managementFrame->addr2()
                 << " with SSID " << managementFrame->ssid() << endl;
        }
        countDevice(stationAddress, signalStrength, ssidProbe, "", "", vector<int>(), vector<float>(), Dot11::MANAGEMENT);
    }
    else if (false && debugMode && managementFrame->subtype() != 8 &&
             managementFrame->subtype() != 4 &&
             managementFrame->subtype() != 5 &&
             managementFrame->subtype() != 12)
    {
        cout << "!Mngmnt frame  -> mac " << managementFrame->addr2()
             << " subtype " << (int)managementFrame->subtype() << endl;
    }
}

void PacketManager::registerControl(Dot11Control *controlFrame,
                                    double signalStrength)
{
    if (showPackets)
    {
        cout << "Control frame  -> " << controlFrame->addr1() << " subtype "
             << (int)controlFrame->subtype() << endl;
    }

    mac address = controlFrame->addr1();
    countDevice(address, signalStrength, "", "", "", vector<int>(), vector<float>(), Dot11::CONTROL);
}

void PacketManager::registerData(Dot11Data *dataFrame, double signalStrength)
{
    if (showPackets)
    {
        mac stationAddress = getStationMAC(dataFrame);
        cout << "Data frame     -> " << stationAddress << " subtype "
             << (int)dataFrame->subtype() << endl;
    }

    mac stationAddress = getStationMAC(dataFrame);
    countDevice(stationAddress, signalStrength, "", "", "", vector<int>(), vector<float>(), Dot11::DATA);
}

int main(int argc, char *argv[])
{
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
    if (!sudo)
    {
        cout << "You must run this program as root!" << endl;
    }

    // Get config from backend
    int secondsPerWindow;
    try
    {
        json backendConfig = getJSON(HOSTNAME + "/v1/config");
        secondsPerWindow = backendConfig["seconds_per_window"];
    }
    catch (UnavailableBackendException &e)
    {
        cout << "Could not connect with backend to get the configuration!"
             << endl
             << "Setting seconds_per_window to 60!" << endl;
        secondsPerWindow = 60;
    }

    // Print important information
    cout << "-----------------------" << endl;
    cout << "Capture device: " << interface << endl;
    cout << "Device ID: " << deviceID << endl;
    if (debugMode)
        cout << "Debug mode enabled!" << endl;
    if (disableUpload)
        cout << "UPLOAD TO BACKEND DISABLED!" << endl;
    cout << "Seconds per window: " << secondsPerWindow << endl;
    cout << "-----------------------" << endl;

    cout << "Enabling monitor mode in interface " << interface << "..." << endl;
    if (!is_monitor_mode(interface))
    {
        // Try to set in monitor mode
        set_monitor_mode(interface);
        if (!is_monitor_mode(interface))
        {
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

    PacketManager *packetManager = new PacketManager(
        disableUpload, deviceID, showPackets, secondsPerWindow);

    while (true)
    {
        Packet pkt = sniffer.next_packet();
        packetManager->registerFrame(pkt);
    }
}
