import pandas as pd
import plotly.express as px
import datetime as dt
import requests
from datetime import datetime
import sys
import plotly.graph_objs as go

# Read data from file
real = pd.read_csv(sys.argv[1], header=0, sep=",")

# Multiply the count column in dataset real
real["count"] = real["count"] * 2.2


par = {
    "from_time": datetime.fromisoformat(min(real["date"])).isoformat('T')+"+02:00",
    "to_time": datetime.fromisoformat(max(real["date"])).isoformat('T')+"+02:00",
}

print(par)

# Get the data from the api
resp = requests.get("https://tfg-api.raporpe.dev/v1/historic", params=par)

rooms = resp.json()["rooms"]

series = []
for room_name in rooms:
    if room_name != "entrada":
        continue
    df = pd.DataFrame([(k, v) for k, v in rooms[room_name].items()], columns=["date", "count"])
    series.append((df, room_name))

for s in series:
    obtained = s[0]

    # Create empty plotly figure
    fig = go.Figure()

    # Plot obtained data in plotly using scatter
    fig.add_trace(go.Scatter(x=obtained["date"], y=obtained["count"], name="Aproximated"))

    # Add title to the figure
    fig.update_layout(title_text="Approximated vs Real")

    # Plot the real data in plotly in magenta colour
    fig.add_trace(go.Scatter(x=real["date"], y=real["count"], name="Real", line=dict(color="magenta")))


    fig.show()