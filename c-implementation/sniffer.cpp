#include <stdio.h>
#include <string>
#include <iostream>
#include <set>
#include <tins/tins.h>


using namespace Tins;
using namespace std;

int main() {

    SnifferConfiguration config;
    config.set_promisc_mode(true);

    Sniffer sniffer("wlan1", config);

    printf("test\n");

    int counter = 0;

    while(true) {
        Packet pkt = sniffer.next_packet();
        Dot11 dot11 = pkt.pdu()->rfind_pdu<Dot11>();

        string addr1 = dot11.addr1().to_string();
        int type = dot11.type();

        cout << addr1 + ", type " << type << endl;

        counter++;

        if(counter % 100 == 0) cout << "------" << counter << "------" << endl;

    }




}