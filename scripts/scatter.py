import psycopg2
import plotly.express as px
import pandas as pd
import time
import datetime
from tqdm import tqdm
import plotly.graph_objs as go

pd.set_option("display.max_rows", None, "display.max_columns", None)

print("Connecting to db...")

conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
cur = conn.cursor()

print("Connected to db!")

pr_vb_query = """
    with pre as (select station_mac, DATE_TRUNC('hour', to_timestamp(time)) as hour, count(*) from probe_request_frames where intent is null and station_mac_vendor is not null group by (station_mac, hour))
        select station_mac, count(*), to_timestamp(avg(extract(epoch from hour))) as time from pre group by (station_mac);
"""
pr_ve_query = """
    with pre as (select station_mac, DATE_TRUNC('hour', to_timestamp(time)) as hour, count(*) from probe_request_frames where intent is not null and station_mac_vendor is not null group by (station_mac, hour))
        select station_mac, count(*), to_timestamp(avg(extract(epoch from hour))) as time from pre group by (station_mac);
"""
pr_fb_query = """
    with pre as (select station_mac, DATE_TRUNC('hour', to_timestamp(time)) as hour, count(*) from probe_request_frames where intent is null and station_mac_vendor is null group by (station_mac, hour))
        select station_mac, count(*), to_timestamp(avg(extract(epoch from hour))) as time from pre group by (station_mac);
"""

pc_fe_query = """
    with pre as (select station_mac, DATE_TRUNC('hour', to_timestamp(time)) as hour, count(*) from probe_request_frames where intent is not null and station_mac_vendor is null group by (station_mac, hour))
        select station_mac, count(*), to_timestamp(avg(extract(epoch from hour))) as time from pre group by (station_mac);
     """

print("Downloading data...")


pr_vb = pd.read_sql_query(pr_vb_query, conn)
pr_ve = pd.read_sql_query(pr_ve_query, conn)
pr_fb = pd.read_sql_query(pr_fb_query, conn)
pr_fe = pd.read_sql_query(pc_fe_query, conn)

print("Transforming data...")


pr_vb["type"] = "vb"
pr_ve["type"] = "ve"
pr_fb["type"] = "fb"
pr_fe["type"] = "fe"

pr = pd.concat([pr_vb, pr_ve, pr_fb, pr_fe])

pr_vb_c = pr_vb["count"].value_counts().to_frame()
pr_ve_c = pr_ve["count"].value_counts().to_frame()
pr_fb_c = pr_fb["count"].value_counts().to_frame()
pr_fe_c = pr_fe["count"].value_counts().to_frame()



pr_vb_c = pr_vb_c.reset_index().rename({'index':'number'}, axis = 'columns')
pr_ve_c = pr_ve_c.reset_index().rename({'index':'number'}, axis = 'columns')
pr_fb_c = pr_fb_c.reset_index().rename({'index':'number'}, axis = 'columns')
pr_fe_c = pr_fe_c.reset_index().rename({'index':'number'}, axis = 'columns')


pr_vb_c["type"] = "vb"
pr_ve_c["type"] = "ve"
pr_fb_c["type"] = "fb"
pr_fe_c["type"] = "fe"

pr_c = pd.concat([pr_vb_c, pr_ve_c, pr_fb_c, pr_fe_c])

print("Making scatter...")

pr_c = pr_c[pr_c["number"] < 70]


fig = px.scatter(pr, x="time", y="count", color="type",
 hover_data=['station_mac'], log_y=True,
 marginal_y="box", title="Mac address ditribution")
fig.show()

fig2 = px.bar(pr_c, x="number", y="count", log_y=True, color="type")
fig2.show()
