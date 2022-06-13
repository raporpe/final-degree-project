from distutils.log import debug
import psycopg2
import plotly.express as px
import pandas as pd
import os

device_id = "raspberry-1"
pd.set_option("display.max_rows", None, "display.max_columns", None)

print("Connecting to db...")

# Get the database password from the environment variable
db_password = os.environ['DB_PASSWORD']

conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=" + db_password)
cur = conn.cursor()

print("Connected to db!")

query = "select * from ocupation order by time"

print("Downloading data...")

df = pd.read_sql_query(query, conn)

print(df)

fig = px.line(df, x="time", y="count", color="device_id")
fig.show()