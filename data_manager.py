from email.quoprimime import body_length
from datetime import datetime
import time
import requests
import logging

API_ENDPOINT = "http://tfg-server.raporpe.dev:2000/v1/upload"
DEVICE_ID = "raspberry-1"
UPLOAD_PERIOD = 10
SEND_DATA_TO_BACKEND = True

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
        self.probe_request_frames = []
        self.beacon_frames = []
        self.data_frames = []
        self.control_frames = []
        self.management_frames = []
        self.probe_response_frames = []

    def register_probe_request_frame(self, station_mac, frequency, power, intent=None):

        self.probe_request_frames.append(
            {
                "station_mac": station_mac,
                "intent": intent if intent != "" else None,
                "time": datetime.now().isoformat(),
                "frequency": frequency,
                "power": power,
            }
        )

        self._send_data()

    def register_probe_response_frame(self, bssid, ssid, station_mac, frequency, power):

        self.probe_response_frames.append(
            {
                "bssid": bssid,
                "ssid": ssid,
                "station_mac": station_mac,
                "time": datetime.now().isoformat(),
                "frequency": frequency,
                "power": power,
            }
        )

        self._send_data()

    def register_beacon_frame(self, bssid, ssid: bytes, frequency):
        
        # Do not register hidden ssid's
        # Hidden ssids container N null characters
        if b'\x00' in ssid:
            return

        beacon = {
            "bssid": bssid,
            "ssid": ssid.decode(),
            "frequency": frequency,
        }

        # If already present, do not insert
        if beacon in self.beacon_frames:
            return

        self.beacon_frames.append(beacon)

    def register_data_frame(self, bssid, station_mac, subtype, frequency, power):

        if not self._validate_mac(bssid) or not self._validate_mac(station_mac):
            return

        self.data_frames.append(
            {
                "bssid": bssid,
                "station_mac": station_mac,
                "time": datetime.now().isoformat(),
                "subtype": subtype,
                "frequency": frequency,
                "power": power,
            }
        )

        self._send_data()

    def register_control_frame(self, addr1, addr2, addr3, addr4, subtype, frequency, power):

        self.control_frames.append(
            {
                "addr1": addr1,
                "addr2": addr2,
                "addr3": addr3,
                "addr4": addr4,
                "time": datetime.now().isoformat(),
                "subtype": str(subtype),
                "frequency": frequency,
                "power": power,
            }
        )

    def register_management_frame(self, addr1, addr2, addr3, addr4, subtype, frequency, power):

        self.management_frames.append(
            {
                "addr1": addr1,
                "addr2": addr2,
                "addr3": addr3,
                "addr4": addr4,
                "time": datetime.now().isoformat(),
                "subtype": str(subtype),
                "frequency": frequency,
                "power": power,
            }
        )

    # Sends the data to the backend

    def _send_data(self):

        # Only send data every N seconds
        if time.time() - self.last_upload_time > UPLOAD_PERIOD:
            print("Passed {} seconds. Uploading data to database.".format(
                UPLOAD_PERIOD))

            json = {
                "device_id": DEVICE_ID,
                "probe_request_frames": self.probe_request_frames,
                "probe_response_frames": self.probe_response_frames,
                "beacon_frames": self.beacon_frames,
                "data_frames": self.data_frames,
                "control_frames": self.control_frames,
                "management_frames": self.management_frames,
            }

            print("Sending data to backend: {probes} probe requests and {beacons} beacons"
                  .format(probes=len(json["probe_request_frames"]), beacons=len(json["beacon_frames"])))

            # Send data to backend in the post payload in json format
            if SEND_DATA_TO_BACKEND:
                requests.post(API_ENDPOINT, json=json)

            print("Uploaded data to backend")

            # Reset the state
            self.last_upload_time = time.time()
            self.probe_request_frames = []
            self.beacon_frames = []
            self.data_frames = []
            self.control_frames = []
            self.management_frames = []
            self.probe_response_frames = []

    def _validate_mac(self, mac):
        return (mac != None
                and mac != "ff:ff:ff:ff:ff:ff"
                and mac != "00:00:00:00:00:00")
