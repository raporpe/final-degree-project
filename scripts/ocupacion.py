from hashlib import new
import psycopg2
import plotly.express as px
import pandas as pd
import datetime as dt
from tqdm import tqdm


class MacManager:

    def __init__(self):
        self.mac_memory = dict()

    def add_macs(self, macs: set):
        mac_memory = self.mac_memory.copy()
        for mac_mem in mac_memory:
            # Si la mac se ha detectado
            if mac_mem in macs:
                if self.mac_memory[mac_mem]["fail"] == 0:
                    self.mac_memory[mac_mem]["row"] += 1
                else:  
                    self.mac_memory[mac_mem]["row"] = 0
                    self.mac_memory[mac_mem]["fail"] = 0
                macs.remove(mac_mem)
            # Si la mac no se ha detectado
            elif mac_mem not in macs:
                if self.mac_memory[mac_mem]["fail"] < 7:
                    self.mac_memory[mac_mem]["fail"] += 1       
                elif self.mac_memory[mac_mem]["fail"] < 15:
                    self.mac_memory[mac_mem]["row"] = 0
                    self.mac_memory[mac_mem]["fail"] += 1

        for mac in macs:
            self.mac_memory[mac] = {
                "row": 1,
                "fail": 0
            }

    def get_current_macs(self):
        count = 0
        for i in self.mac_memory:
            if self.mac_memory[i]["row"] > 1:
                count += 1
        return count

class Mac:
    def __init__(self, addr):
        self.addr = addr



# Crete variable of date
start_time = dt.datetime.fromisoformat("2022-02-24T12:00:00")

# Sum 14 days to start_time
end_time = start_time + dt.timedelta(days=14)


slice_size = 60
slices_to_active = 5

print("Connecting to db...")

conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
cur = conn.cursor()


probe_request_query = "select * from \"old-tfg\".probe_request_frames where time < '{e}' and time > '{s}'".format(e=end_time.isoformat(), s=start_time.isoformat())
print(probe_request_query)
probe_request_df = pd.read_sql_query(probe_request_query, conn)

dates = []
# Append all the dates between start_time and end_time in increments of slice_size
for i in range(0, int((end_time - start_time).total_seconds()), slice_size):
    dates.append(start_time + dt.timedelta(seconds=i))


print(dates)

# Initialize the mac address manager
mac_manager = MacManager()

new_mac_addresses = []
new_mac_addresses_cumulative = []
times = []

# slice_size to seconds in datetime object
slice_size_dt = dt.timedelta(seconds=slice_size)

mac_set = set()

for t in tqdm(dates):

    lst = probe_request_df.loc[
        (probe_request_df["time"] > (t - slice_size_dt)) &
        (probe_request_df["time"] < t),
    ]["station_mac"].tolist()

    # Get the number of elements in lst that are in the set mac_set
    s = set(lst)
    new_macs = len(s) - len(s.intersection(mac_set))


    # Add lst to the set mac_set
    mac_set.update(lst)

    new_mac_addresses.append(new_macs)
    new_mac_addresses_cumulative.append(len(mac_set))
    times.append(t)


data = {
    "time": times,
    "new MAC addresses": new_mac_addresses,
    "new MAC addresses cumulative": new_mac_addresses_cumulative
}
df = pd.DataFrame(data)

#fig = px.line(df, x="time", y=["new MAC addresses", "new MAC addresses cumulative"])
fig = px.line(df, x="time", y="new MAC addresses")


# Add more ticks in the figure x axis
fig.update_xaxes(dtick="1d")


fig.show()




