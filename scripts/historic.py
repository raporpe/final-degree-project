import pandas as pd
import plotly.express as px
import datetime as dt
import requests
import time

# Get the data from the api
p1 = {
    "from_time": dt.datetime.now().replace(hour=0, minute=0, second=0, microsecond=0).isoformat() + "Z",
    "to_time": dt.datetime.now().isoformat() + "Z"
}

p2 = {
    "from_time": "2022-05-19T00:00:00Z",
    "to_time": dt.datetime.now().isoformat() + "Z"
}

resp = requests.get("https://tfg-api.raporpe.dev/v1/historic", params=p1)


rooms = resp.json()["rooms"]


series = []
for room_name in rooms:
    df = pd.DataFrame([(k, v) for k, v in rooms[room_name].items()], columns=["hora", "ocupacion"])
    series.append((df, room_name))

print(series[0])

for s in series:
    fig = px.line(s[0], x='hora', y="ocupacion", title=s[1])
    fig.show()
    time.sleep(5)