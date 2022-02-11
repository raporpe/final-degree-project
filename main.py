from scapy.all import *
import traceback
from data_manager import DataManager


def packet_handler(pkt):
    if pkt.haslayer(Dot11ProbeReq):

        print("Packet with MAC {pkt.addr2}, power {pkt.dBm_AntSignal} and SSID {ssid} ".format(show=pkt.show(dump=True), pkt=pkt, ssid=pkt.info.decode()))
        
        manager = DataManager()
        manager.register_probe_request(station_bssid=pkt.addr2, intent=pkt.info.decode(), power=pkt.dBm_AntSignal)


    elif pkt.haslayer(Dot11Beacon):
        manager = DataManager()
        manager.register_beacon(pkt.addr3, pkt.info.decode())
    

        

def start_sniffer():
    try:
        sniff(iface="wlan1mon", prn=packet_handler, store=0)
    except Exception as e:
        print("--------")
        traceback.print_exc()
        start_sniffer()

if __name__ == "__main__":
    start_sniffer()
