from audioop import add
from scapy.all import Dot11, Dot11ProbeReq, Dot11Beacon, sniff
import traceback
from data_manager import DataManager
from mac_vendor_lookup import MacLookup, VendorNotFoundError


# Get the mac address of the wireless mobile device
# that we are interested in targeting
def get_mac_from_pkt(pkt):

    # If the address 4 has meaning, then the packet is from a wireless bridge
    # This means that no mobile wireless device is involved
    #print("Meaning " + str(pkt[Dot11].address_meaning(4)))
    if pkt[Dot11].address_meaning(4) != None:
        return None


    addr_matrix = [
        {
            "meaning": pkt[Dot11].address_meaning(1),
            "addr": pkt.addr1
        },
        {
            "meaning": pkt[Dot11].address_meaning(2),
            "addr": pkt.addr2
        },
        {
            "meaning": pkt[Dot11].address_meaning(3),
            "addr": pkt.addr3
        }
    ]

    # Detect the BSSID
    bssid = None
    for addr in addr_matrix:
        if "BSSID" in addr["meaning"]:
            bssid = addr["addr"]

    # Get addr where TA=SA
    # This means that the Transmission Address = Source Address
    for addr in addr_matrix:
        if "TA=SA" in addr["meaning"]:
            print("TA=SA")
            return addr["addr"]

    # Is the previous one was not the case, then RA=DA is the one
    # Where Receiving Address = Destination Address
    for addr in addr_matrix:
        if "RA=DA" in addr["meaning"] and addr["addr"] != bssid:
            print("RA=DA")
            return addr["addr"]

    
    return None



def packet_handler(pkt):
    if pkt.haslayer(Dot11):

        # Probe requests
        if pkt.type == 0 and pkt.subtype == 4:

            print("Packet with MAC {pkt.addr2}, power {pkt.dBm_AntSignal} and SSID {ssid} ".format(
                show=pkt.show(dump=True), pkt=pkt, ssid=pkt.info.decode()))

            manager = DataManager()
            manager.register_probe_request(
                station_bssid=pkt.addr2, intent=pkt.info.decode(), power=pkt.dBm_AntSignal)

        # Beacons
        elif pkt.type == 0 and pkt.subtype == 8:
            manager = DataManager()
            manager.register_beacon(pkt.addr3, pkt.info.decode())

        # Data
        elif pkt.type == 2:
            pass

            #print(str(pkt.show()) + str(pkt.addr1))


def start_sniffer():
    try:
        sniff(iface="wlan1mon", prn=packet_handler, store=0)
    except Exception as e:
        print("--------")
        traceback.print_exc()
        start_sniffer()


if __name__ == "__main__":
    start_sniffer()
