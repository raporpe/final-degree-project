#ifndef HELPERS_H
#define HELPERS_H

#include "main.h"
#include "json.hpp"
#include <tins/tins.h>
#include <regex>
#include <thread>
#include <stdexcept>

using json = nlohmann::json;

int getCurrentTime();

mac getStationMAC(Tins::Dot11Data *frame);

bool isMacFake(mac address);

bool isMacValid(mac address);

json postJSON(string url, json j);

json getJSON(string url);

void channel_switcher(string interface);

void set_monitor_mode(string interface);

bool is_monitor_mode(string interface);

size_t curlWriteCallback(void *contents, size_t size, size_t nmemb, std::string *s);

struct UnavailableBackendException : public exception
{
	const char * what () const throw ()
    {
    	return "The backend is not available in this moment.";
    }
};

#endif