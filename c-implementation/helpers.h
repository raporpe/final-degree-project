#ifndef HELPERS_H
#define HELPERS_H

#include <tins/tins.h>
#include "main.h"

int getCurrentTime();

mac getStationMAC(Tins::Dot11Data *frame);

bool isMacFake(mac address);

bool isMacValid(mac address);

#endif