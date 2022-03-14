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

query = "select * from ocupation order by time"

print("Downloading data...")

df = pd.read_sql_query(query, conn)

print(df)

fig = px.line(df, x="time", y="count")
fig.show()