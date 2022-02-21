import psycopg2
import plotly.express as px
import pandas as pd
import time
import datetime
from tqdm import tqdm

def deacumulate(a):
    ret = [0]
    for idx in range(len(a)-1):
        ret.append(a[idx+1] - a[idx])
    return ret


conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
cur = conn.cursor()

print("Getting data from db....")

probe_request_query = "select * from probe_request_frames"
probe_response_query = "select * from probe_response_frames"
data_query = "select * from data_frames"
probe_request_df = pd.read_sql_query(probe_request_query, conn)
probe_response_df = pd.read_sql_query(probe_response_query, conn)
data_df = pd.read_sql_query(data_query, conn)

# Calculate real macs increse graph
start_time = 1644505200
end_time = int(time.time())
probe_request_real_macs_cumulative = []
probe_request_fake_macs_cumulative = []
probe_response_real_macs_cumulative = []
data_cumulative = []
probe_request_and_probe_response_real_macs = []
all_real_cumulative = [0 for i in range(start_time, end_time, 3600)]
probe_request_and_probe_response_real_macs = [0 for i in range(start_time, end_time, 3600)]
unix_time = [i for i in range(start_time, end_time, 3600)]

print("Calculating graphs...")

print(probe_response_df.loc[probe_response_df['station_mac_vendor'] == None, "station_mac_vendor"])

# Real macs from probe responses
for idx, t in tqdm(enumerate(range(start_time, end_time, 3600))):
    q = probe_response_df
    q = q.loc[
        (q['station_mac_vendor'].notna()) &
        (q['frequency'] < 3000) & 
        (q['time'] < t),]
    q = len(pd.unique(q["station_mac"]))

    all_real_cumulative[idx] += q
    probe_request_and_probe_response_real_macs[idx] += q
    probe_response_real_macs_cumulative.append(q)

# Real macs from data frames responses
for idx, t in tqdm(enumerate(range(start_time, end_time, 3600))):
    q = data_df
    q = q.loc[
        (q['station_mac_vendor'].notna()) & 
        (q['frequency'] < 3000) & 
        (q['time'] < t),]

    q = len(pd.unique(q["station_mac"]))
    all_real_cumulative[idx] += q
    probe_request_and_probe_response_real_macs[idx] += q

    data_cumulative.append(q)


# Real macs
for idx, t in tqdm(enumerate(range(start_time, end_time, 3600))):
    q = probe_request_df
    q = q.loc[
        (q['station_mac_vendor'].notna()) & 
        (q['frequency'] < 3000) & 
        (q['time'] < t)
        ,]


    q = len(pd.unique(q["station_mac"]))
    all_real_cumulative[idx] += q

    probe_request_real_macs_cumulative.append(q)

# Fake macs from probe request
for idx, t in tqdm(enumerate(range(start_time, end_time, 3600))):
    q = probe_request_df
    q = q.loc[(q['time'] < t) &
        (q['frequency'] < 3000)
    ,]

    q = len(pd.unique(q["station_mac"]))
    probe_request_fake_macs_cumulative.append(q)

# Real macs from probe requests + probe responses
probe_request_and_probe_response_real_macs = list(probe_request_and_probe_response_real_macs)

# Real macs from probe responses + probe requests + data frames
all_real_cumulative = list(all_real_cumulative)

probe_request_real_macs = deacumulate(probe_request_real_macs_cumulative)
probe_request_fake_macs = deacumulate(probe_request_fake_macs_cumulative)
probe_response_real_macs = deacumulate(probe_response_real_macs_cumulative)
df = pd.DataFrame({
    "time": unix_time,
    "probe_request_fake_macs_cumulative": probe_request_fake_macs_cumulative,
    "probe_request_fake_macs": probe_request_fake_macs,
    "probe_request_real_macs_cumulative": probe_request_real_macs_cumulative,
    "probe_request_real_macs": probe_request_real_macs,
    "probe_response_real_macs": probe_response_real_macs,
    "probe_response_real_macs_cumulative": probe_response_real_macs_cumulative,
    "data_cumulative": data_cumulative,
    "probe_request_and_probe_response_real_macs": probe_request_and_probe_response_real_macs,
    "all_real_cumulative": all_real_cumulative
})

df.time = df.time.apply(lambda d: datetime.datetime.fromtimestamp(
    int(d)).strftime('%d %a - %Hh'))

fig = px.line(df, x="time", 
                y=["probe_request_fake_macs_cumulative", "probe_request_fake_macs",
                    "probe_request_real_macs_cumulative", "probe_request_real_macs",
                    "probe_response_real_macs","probe_response_real_macs_cumulative",
                    "data_cumulative", "probe_request_and_probe_response_real_macs",
                    "all_real_cumulative"],
                title="Detected unique macs")

dates = df["time"].to_list()
dates = set([i[:6] for i in dates])
print(dates)

for date in dates:
    fig.add_vrect(x0=date+" - 00h", x1=date+" - 08h",
                  annotation_text="night", annotation_position="top left",
                  fillcolor="blue", opacity=0.25, line_width=0)

fig.show()