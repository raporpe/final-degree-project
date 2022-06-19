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

api = pd.DataFrame([(k, v) for k, v in biblioteca.items()], columns=["date", "count"])

# Multiply the count column in dataset obtained
api["count"] = api["count"]


# When all the csv files have been generated
if len(sys.argv) > 2 and sys.argv[2] == "abstract":
    # Read the csv from the third argument
    csv = pd.read_csv(sys.argv[3])

    # Multiply the optics columns in the csv by a factor
    csv["optics"] = csv["optics"] * 0.25

    # Delete the last five rows of the csv
    csv = csv[:-1]
    real = real[:-1*6]

    # Plotly empty figure
    fig = go.Figure()

    # Plot the information in the csv
    fig.add_trace(go.Scatter(x=csv['date'], y=csv['optics'], name='t-SNE + OPTICS'))
    fig.add_trace(go.Scatter(x=csv['date'], y=csv['simple'], name='Vendor tags clustering'))
    fig.add_trace(go.Scatter(x=real['date'], y=real['count'], name='Real'))

    # Set the title
    fig.update_layout(title_text="Clustering algorithms comparison")

    # Set the x axis
    fig.update_xaxes(title_text="Date")

    # Set the y axis
    fig.update_yaxes(title_text="People")

    # Fix box at the bottom of the generated PDF
    import plotly.io as pio   
    pio.kaleido.scope.mathjax = None

    # Save figure as pdf
    fig.write_image(sys.argv[3].replace(".csv", "") + ".pdf", width=1750/1.5, height=800/1.5, format="pdf")

    fig.show()

else:
    # Create empty plotly figure
    fig = go.Figure()

    # Plot obtained data in plotly using scatter
    fig.add_trace(go.Scatter(x=api["date"], y=api["count"], name="Aproximated"))

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
            "date": api["date"],
            "optics": existing_csv["optics"],
            "simple": api["count"],
        })

        # Save the frame in a csv file
        df.to_csv(sys.argv[3], index=False)
