#ifndef HELPERS_H
#define HELPERS_H

#include <tins/tins.h>
#include <json.hpp>
#include "main.h"

using json = nlohmann::json;

int getCurrentTime();

mac getStationMAC(Tins::Dot11Data *frame);

bool isMacFake(mac address);

bool isMacValid(mac address);

void postJSON(string url, json j);

json getJSON(string url);


#endif