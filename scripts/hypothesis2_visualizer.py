from turtle import width
import pandas as pd
import plotly.express as px
import datetime as dt
import requests
from datetime import datetime
import sys
import plotly.graph_objs as go

par = {
    "from_time": "2022-05-01T00:00:00+02:00",
    "to_time": "2022-06-03T00:00:00+02:00",
}

print(par)

# Force historic recalculation
if len(sys.argv) > 2 and sys.argv[2] == "true":
    requests.get("https://tfg-api.raporpe.dev/v1/historic-recalc", params=par)

# Get the data from the api
resp = requests.get("https://tfg-api.raporpe.dev/v1/historic", params=par)


rooms = resp.json()["rooms"]

# Delete every room except makerspace
rooms = {k: v for k, v in rooms.items() if k == "makerspace"}


print(rooms)

# Convert from json response to dataframe format
series = []
for room_name in rooms:
    df = pd.DataFrame([(k, v) for k, v in rooms[room_name].items()], columns=["date", "count"])
    series.append((df, room_name))

for s in series:
    obtained = s[0]

    # Only keep the dates from may 27 01:00 to may 29
    obtained = obtained[obtained["date"].between("2022-05-27T02:00:00+02:00", "2022-05-29T05:30:00+02:00")]

    # Create empty plotly figure
    fig = go.Figure()

    # Set figure title to the room name
    fig.update_layout(title_text=s[1])

    # Set x and y axis titles
    fig.update_xaxes(title_text="Date")
    fig.update_yaxes(title_text="Count")

    # y axis starts at 0 in plotly
    fig.update_yaxes(autorange=False, range=[0, max(obtained["count"])*1.1])

    # Set empty graph title
    fig.update_layout(title_text="")

    # Plot obtained data in plotly using scatter
    fig.add_trace(go.Scatter(x=obtained["date"], y=obtained["count"], name="Aproximated"))

    # Disable MathMenu ajax in pdf
    import plotly.io as pio   
    pio.kaleido.scope.mathjax = None

    fig.show()

    # Save the figure in pdf file
    fig.write_image("data1_results.pdf", width=2600/1.6, height=700/1.4)