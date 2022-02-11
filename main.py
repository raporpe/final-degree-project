from http.client import REQUEST_ENTITY_TOO_LARGE
from scapy.all import Dot11ProbeReq, sniff
import traceback
import requests
import time

API_ENDPOINT = "http://tfg-server.raporpe.dev:2000/v1/upload"
DEVICE_ID = "raspberry-1"
UPLOAD_PERIOD = 10

counter = 0

def packet_handler(pkt):
    if pkt.haslayer(Dot11ProbeReq):
        global counter, f
        counter += 1
        print("Packet with MAC {pkt.addr2}, power {pkt.dBm_AntSignal} and SSID {ssid} ".format(show=pkt.show(dump=True), pkt=pkt, ssid=pkt.info.decode()))
        add_data(station_bssid=pkt.addr2, intent=pkt.info.decode(), power=pkt.dBm_AntSignal)
        if counter % 100 == 0:
            print(counter)
        

last_time = 0
current_data = []
def add_data(station_bssid, power, intent=None):
    global current_data
    current_data.append(
		{
			"station_bssid": station_bssid,
            "intent": intent,
            "time": int(time.time()),
            "power": power
		}
    )

    global last_time
    if time.time() - last_time > UPLOAD_PERIOD:
        print("Passed {} seconds. Uploading data to database.".format(UPLOAD_PERIOD))
        send_data(current_data)
        last_time = time.time()
        current_data = []


def send_data(probe_requests):

    json = 	{ 
        "device_id": DEVICE_ID,
        "probe_requests": probe_requests
    }

    print("Inserting {} records".format(len(json["probe_requests"])))
    requests.post(API_ENDPOINT, json=json)
    print("Uploaded data to backend")


def start_sniffer():
    try:
        sniff(iface="wlan1mon", prn=packet_handler, store=0)
    except Exception as e:
        print("--------")
        traceback.print_exc()
        start_sniffer()

if __name__ == "__main__":
    start_sniffer()
