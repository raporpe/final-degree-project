import psycopg2
import plotly.express as px
import pandas as pd
import datetime as dt
from tqdm import tqdm
import os


# Set the start time
start_time = dt.datetime.fromisoformat("2022-02-24T12:00:00")

# Set the end time 14 days later
end_time = start_time + dt.timedelta(days=14)

# The window size is of 60 seconds
window_size = 60

print("Connecting to db...")

# Import the database password from the environment variable
db_password = os.environ['DB_PASSWORD']

# Connect to the database
conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=" + db_password)
cur = conn.cursor()

# Get all the Probe Requests from the database
probe_request_query = "select * from \"old-tfg\".probe_request_frames where time < '{e}' and time > '{s}'".format(e=end_time.isoformat(), s=start_time.isoformat())
print(probe_request_query)
probe_request_df = pd.read_sql_query(probe_request_query, conn)

dates = []
# Generate a list with all the starting times of the windows
for i in range(0, int((end_time - start_time).total_seconds()), window_size):
    dates.append(start_time + dt.timedelta(seconds=i))


new_mac_addresses = []
new_mac_addresses_cumulative = []
times = []

# slice_size to seconds in datetime object
slice_size_dt = dt.timedelta(seconds=window_size)

mac_set = set()

for t in tqdm(dates):

    # Divide the dataframe into slices of 60 seconds
    probe_request_df_slice = probe_request_df[(probe_request_df["time"] >= t) & (probe_request_df["time"] < t + slice_size_dt)]

    # Get the unique mac addresses from the slice
    mac_addresses = set(probe_request_df_slice["station_mac"])

    # Get the number of mac addresses that are not in mac_set
    new_mac_addresses.append(len(mac_addresses - mac_set))

    # Add the mac_addresses to the mac_set
    mac_set = mac_set | mac_addresses

    times.append(t)


data = {
    "time": times,
    "new MAC addresses": new_mac_addresses,
}
df = pd.DataFrame(data)


fig = px.line(df, x="time", y="new MAC addresses")

# Add more ticks in the figure x axis
fig.update_xaxes(dtick="1d")

# Disable mathjax
# import plotly.io as pio   
# pio.kaleido.scope.mathjax = None

# Save the figure as pdf

fig.write_pdf("ocupacion.pdf", width=4000, height=1000)


fig.show()




