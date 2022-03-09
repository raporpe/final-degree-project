import psycopg2
import plotly.express as px
import pandas as pd
import datetime
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



start_time = 1644505200 + 3600*24*7
end_time = start_time + 3600*24*14
slice_size = 60
slices_to_active = 5

print("Connecting to db...")

conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
cur = conn.cursor()


probe_request_query = "select * from probe_request_frames where time < {e} and time > {s}".format(e=end_time, s=start_time)
probe_request_df = pd.read_sql_query(probe_request_query, conn)

slices = [i for i in range(start_time, end_time, slice_size)]

# Initialize the mac address manager
mac_manager = MacManager()

series = []
times = []

for t in tqdm(slices):

    lst = probe_request_df.loc[
        (probe_request_df["time"] > (t - slice_size)) &
        (probe_request_df["time"] < t),
    ]["station_mac"].tolist()

    mac_manager.add_macs(set(lst))
    n = mac_manager.get_current_macs()
    series.append(n)
    times.append(t)


data = {
    "time": times,
    "addresses":series
}
df = pd.DataFrame(data)
df.time = df.time.apply(lambda d: datetime.datetime.fromtimestamp(
    int(d)).strftime('%d %a - %Hh:%M'))

fig = px.line(df, x="time", y="addresses")
fig.show()




