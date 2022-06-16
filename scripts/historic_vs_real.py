import pandas as pd
import plotly.express as px
import datetime as dt
import requests
from datetime import datetime
import sys
import plotly.graph_objs as go

# Read data from file
real = pd.read_csv(sys.argv[1], header=0, sep=",")


par = {
    "from_time": datetime.fromisoformat(min(real["date"])).isoformat('T')+"+02:00",
    "to_time": datetime.fromisoformat(max(real["date"])).isoformat('T')+"+02:00",
}

print(par)

# Force historic recalculation
if len(sys.argv) > 2 and sys.argv[2] == "true":
    requests.get("https://tfg-api.raporpe.dev/v1/historic-recalc", params=par)

# Get the data from the api
resp = requests.get("https://tfg-api.raporpe.dev/v1/historic", params=par)


biblioteca = resp.json()["rooms"]["entrada"]

df = pd.DataFrame([(k, v) for k, v in biblioteca.items()], columns=["date", "count"])

obtained = df

# Multiply the count column in dataset obtained
obtained["count"] = obtained["count"]

# Create empty plotly figure
fig = go.Figure()

# Plot obtained data in plotly using scatter
fig.add_trace(go.Scatter(x=obtained["date"], y=obtained["count"], name="Aproximated"))

# Add title to the figure
fig.update_layout(title_text="Approximated vs Real")

# Plot the real data in plotly in magenta colour
fig.add_trace(go.Scatter(x=real["date"], y=real["count"], name="Real", line=dict(color="magenta")))


fig.show()


# Also save the data in a csv called as in third argument
if len(sys.argv) > 3:

    # Read the csv file in the third argument
    existing_csv = pd.read_csv(sys.argv[3], header=0, sep=",")


    # Create a pandas with the real data + the count column from obtained
    df = pd.DataFrame({
        "date": obtained["date"],
        "optics": obtained["count"],
        "simple": existing_csv["simple"]
    })

    # Save the frame in a csv file
    df.to_csv(sys.argv[3], index=False)
