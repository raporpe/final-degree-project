import psycopg2
import plotly.express as px
import pandas as pd
import time
import datetime

def deacumulate(a):
    ret = [0]
    for idx in range(len(a)-1):
        ret.append(a[idx+1] - a[idx])
    return ret


conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
cur = conn.cursor()


# Calculate real macs increse graph
start_time = 1644505200
end_time = int(time.time())
real_macs_cumulative = []
quite_real_macs_cumulative = []
fake_macs_cumulative = []
unix_time = [i for i in range(start_time, end_time, 3600)]


for t in range(start_time, end_time, 3600):
    cur.execute("SELECT count(distinct station_mac) FROM probe_request_frames WHERE station_mac_vendor is not null and time < '{}'".format(t))
    number = cur.fetchall()[0][0]
    print(number)
    real_macs_cumulative.append(number)


for t in range(start_time, end_time, 3600):
    cur.execute("""select count(*) from 
                (SELECT distinct station_mac FROM probe_request_frames WHERE station_mac_vendor is not null and time < '{}'
                INTERSECT
                SELECT station_mac FROM probe_request_frames GROUP BY station_mac HAVING COUNT(*) > 1
                ) AS E""".format(t))
    number = cur.fetchall()[0][0]
    print(number)
    quite_real_macs_cumulative.append(number)



for t in range(start_time, end_time, 3600):
    cur.execute("SELECT COUNT(distinct station_mac) FROM probe_request_frames WHERE station_mac_vendor is null and time < '{}'".format(t))
    number = cur.fetchall()[0][0]
    print(number)
    fake_macs_cumulative.append(number)

real_macs = deacumulate(real_macs_cumulative)
fake_macs = deacumulate(fake_macs_cumulative)
quite_real_macs = deacumulate(quite_real_macs_cumulative)

print(len(unix_time))
print(len(fake_macs_cumulative))
print(len(fake_macs))
print(len(real_macs_cumulative))
print(len(real_macs))
print(len(quite_real_macs_cumulative))
print(len(quite_real_macs))

df = pd.DataFrame({
    "time": unix_time,
    "fake_macs_cumulative": fake_macs_cumulative,
    "fake_macs": fake_macs,
    "real_macs_cumulative": real_macs_cumulative,
    "real_macs": real_macs,
    "quite_real_macs_cumulative": quite_real_macs_cumulative,
    "quite_real_macs": quite_real_macs
})

df.time = df.time.apply(lambda d: datetime.datetime.fromtimestamp(
    int(d)).strftime('%d %a - %Hh'))

fig = px.line(df, x="time", 
                y=["fake_macs", "real_macs", "fake_macs_cumulative",
                    "real_macs_cumulative", "quite_real_macs", "quite_real_macs_cumulative"],
                title="Detected unique macs")

dates = df["time"].to_list()
dates = set([i[:6] for i in dates])
print(dates)

for date in dates:
    fig.add_vrect(x0=date+" - 00h", x1=date+" - 08h",
                  annotation_text="night", annotation_position="top left",
                  fillcolor="blue", opacity=0.25, line_width=0)

fig.show()