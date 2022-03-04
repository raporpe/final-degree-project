import psycopg2
import plotly.express as px
import pandas as pd
import time
import datetime
from tqdm import tqdm

SLICE = 3600

def acumulate(a):
    ret = [a[0]]
    cum  = a[0]
    for i in range(1, len(a), 1):
        cum += a[i]
        ret.append(cum)

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
probe_request_real_macs = []
probe_request_fake_macs = []
probe_response_real_macs = []
data_real_macs = []
probe_request_and_probe_response_real_macs = []
all_real = []
unix_time = []

print("Calculating graphs...")

for idx, t in tqdm(enumerate(range(start_time, end_time, SLICE))):
    # Real macs from probe responses
    pr = probe_response_df
    pr = pr.loc[
        (pr['station_mac_vendor'].notna()) &
        (pr['frequency'] < 3000) &
        (pr['time'] > t-SLICE) &
        (pr['time'] < t), ]

    probe_response_real_macs.append(
        len(pd.unique(pr["station_mac"])))

    # Real macs from data frames responses
    d = data_df
    d = d.loc[
        (d['station_mac_vendor'].notna()) &
        (d['frequency'] < 3000) &
        (d['time'] > t-SLICE) &
        (d['time'] < t), ]

    data_real_macs.append(
        len(pd.unique(d["station_mac"])))

    # Real macs
    pre = probe_request_df
    pre = pre.loc[
        (pre['station_mac_vendor'].notna()) &
        (pre['frequency'] < 3000) &
        (pre['time'] > t-SLICE) &
        (pre['time'] < t), ]

    probe_request_real_macs.append(
        len(pd.unique(pre["station_mac"])))

    # Fake macs from probe request
    pref = probe_request_df
    pref = pref.loc[(pref['time'] < t) &
                    (pref['time'] > t-SLICE) &
                    (pref['frequency'] < 3000), ]

    probe_request_fake_macs.append(
        len(pd.unique(pref["station_mac"])))

    # All data: pr + pre + data
    all_real.append(
        len(
            pd.unique(
                pd.concat(
                    [pre["station_mac"], pr["station_mac"], d["station_mac"]])
            )
        )
    )

    # Probe request + probe responses
    probe_request_and_probe_response_real_macs.append(
        len(pd.unique(
                pd.concat([pre["station_mac"], pr["station_mac"]])
            )
        )
    )

    unix_time.append(t)


probe_request_real_macs_cumulative = acumulate(probe_request_real_macs)
probe_request_fake_macs_cumulative = acumulate(probe_request_fake_macs)
probe_response_real_macs_cumulative = acumulate(probe_response_real_macs)
data_real_macs_cumulative = acumulate(data_real_macs)
all_real_cumulative = acumulate(all_real)


df = pd.DataFrame({
    "time": unix_time,
    "probe_request_fake_macs_cumulative": probe_request_fake_macs_cumulative,
    "probe_request_fake_macs": probe_request_fake_macs,
    "probe_request_real_macs_cumulative": probe_request_real_macs_cumulative,
    "probe_request_real_macs": probe_request_real_macs,
    "probe_response_real_macs_cumulative": probe_response_real_macs_cumulative,
    "probe_response_real_macs": probe_response_real_macs,
    "data_cumulative": data_real_macs_cumulative,
    "probe_request_and_probe_response_real_macs": probe_request_and_probe_response_real_macs,
    "all_real_cumulative": all_real_cumulative
})

df.time = df.time.apply(lambda d: datetime.datetime.fromtimestamp(
    int(d)).strftime('%d %a - %Hh'))

fig = px.line(df, x="time",
              y=["probe_request_fake_macs_cumulative", "probe_request_fake_macs",
                    "probe_request_real_macs_cumulative", "probe_request_real_macs",
                    "probe_response_real_macs", "probe_response_real_macs_cumulative",
                    "data_cumulative", "probe_request_and_probe_response_real_macs",
                    "all_real_cumulative"],
              title="Detected unique macs")

print(df["time"])

dates = df["time"].to_list()
dates = set([i[:6] for i in dates])
print(dates)

for date in dates:
    fig.add_vrect(x0=date+" - 00h", x1=date+" - 08h",
                  annotation_text="night", annotation_position="top left",
                  fillcolor="blue", opacity=0.25, line_width=0)

fig.show()
