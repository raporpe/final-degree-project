import psycopg2
import plotly.express as px
import pandas as pd
import numpy as np
import time
import datetime


conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
cur = conn.cursor()


# Calculate real macs increse graph
start_time = 1644505200
end_time = int(time.time())
real_macs = []
fake_macs = []
unix_time = []

for t in range(start_time, end_time, 3600):
    cur.execute("SELECT COUNT(distinct station_mac) FROM probe_request_frames WHERE station_mac_vendor is not null and time < '{}'".format(t))
    real_mac_number = cur.fetchall()[0][0]
    print(real_mac_number)
    real_macs.append(real_mac_number)
    unix_time.append(t)

for t in range(start_time, end_time, 3600):
    cur.execute("SELECT COUNT(distinct station_mac) FROM probe_request_frames WHERE station_mac_vendor is null and time < '{}'".format(t))
    fake_mac_number = cur.fetchall()[0][0]
    print(fake_mac_number)
    fake_macs.append(fake_mac_number)


df = pd.DataFrame({
    "time": unix_time,
    "fake_macs": fake_macs,
    "real_macs": real_macs
})

df.time = df.time.apply(lambda d: datetime.datetime.fromtimestamp(
    int(d)).strftime('%d %a - %Hh'))

tick = []

fig = px.line(df, x="time", y=["fake_macs", "real_macs"], title="Real macs")

dates = df["time"].to_list()
dates = set([i[:6] for i in dates])

for date in dates:
    fig.add_vrect(x0=date+" - 00h", x1=date+" - 08h",
                  annotation_text="night", annotation_position="top left",
                  fillcolor="blue", opacity=0.25, line_width=0)

fig.show()
