import time
import requests
from mac_vendor_lookup import MacLookup, VendorNotFoundError


API_ENDPOINT = "http://tfg-server.raporpe.dev:2000/v1/upload"
DEVICE_ID = "raspberry-1"
UPLOAD_PERIOD = 10

# Metaclass manager for DataManager


class Singleton(type):
    _instances = {}  #  Set of instances of the classes

    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:  # If not instantiated
            cls._instances[cls] = super(
                Singleton, cls).__call__(*args, **kwargs)
        return cls._instances[cls]


class DataManager(metaclass=Singleton):

    def __init__(self):
        self.last_upload_time = time.time()
        self.current_probe_requests = []
        self.current_beacons = []

    def register_probe_request(self, station_bssid, power, intent=None):
        try:
            vendor = MacLookup().lookup(station_bssid)
        except VendorNotFoundError:
            vendor = None

        self.current_probe_requests.append(
            {
                "station_bssid": station_bssid,
                "intent": intent,
                "time": int(time.time()),
                "power": power,
                "vendor": vendor
            }
        )

        self.send_data()

    def register_beacon(self, bssid, ssid):
        beacon = {
            "bssid": bssid,
            "ssid": ssid
        }

        if beacon in self.current_beacons:
            return

        self.current_beacons.append(beacon)

        self.send_data()

    # Sends the data to the backend
    def send_data(self):

        # Only send data every N seconds
        if time.time() - self.last_upload_time > UPLOAD_PERIOD:
            print("Passed {} seconds. Uploading data to database.".format(
                UPLOAD_PERIOD))

            json = {
                "device_id": DEVICE_ID,
                "probe_requests": self.current_probe_requests,
                "beacons": self.current_beacons
            }

            print("Sending data to backend: {probes} probe requests and {beacons} beacons"
                  .format(probes=json["probe_requests"], beacons=json["beacons"]))

            # Send data to backend in the post payload
            requests.post(API_ENDPOINT, json=json)

            print("Uploaded data to backend")

            # Reset the state
            self.last_upload_time = time.time()
            self.current_probe_requests = []
            self.current_beacons = []
