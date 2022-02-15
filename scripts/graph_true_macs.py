import psycopg2
import plotly.express as px
import pandas as pd
import numpy as np
import time
import datetime


conn = psycopg2.connect("host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
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

real_macs = pd.DataFrame({
    "time": unix_time,
    "data": real_macs
})

fake_macs = pd.DataFrame({
    "time": unix_time,
    "data": fake_macs
})

real_macs.time = real_macs.time.apply(lambda d: datetime.datetime.fromtimestamp(int(d)).strftime('%d %a - %Hh'))
fake_macs.time = fake_macs.time.apply(lambda d: datetime.datetime.fromtimestamp(int(d)).strftime('%d %a - %Hh'))


tick = []

fig = px.line(real_macs, x="time", y="data")
fig.write_image("real_macs.png")


fig = px.line(fake_macs, x="time", y="data")
fig.write_image("fake_macs.png")