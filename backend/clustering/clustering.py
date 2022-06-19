# t-SNE is used for dimensionality reduction
from sklearn.manifold import TSNE

# OPTICS is used for clustering
from sklearn.cluster import OPTICS

# Definition of a digested MAC in Pydantic model
from pydantic import BaseModel

# Import logarithms
from math import log

# Import time package
import time 


# The model of a digested MAC to cluster
class DigestedMAC(BaseModel):
    mac: str
    average_signal_strenght: float
    manufacturer: str | None
    oui_id: str
    type_count: list[int]
    presence_record: list[bool]
    ssid_probes: list[str] | None
    ht_capabilities: str | None
    ht_extended_capabilities: str | None
    supported_rates: list[float] | None
    tags: list[int] | None


# FastAPI will be used for receiving data from Go
from fastapi import FastAPI, Request, status
# Import requestvalidationerror
from fastapi.exceptions import RequestValidationError
# Import JSONResponse
from fastapi.responses import JSONResponse
import logging

# Define the single API route for receiving digested MACs
app = FastAPI()

debug = True

@app.post("/cluster")
def receive_digested_macs(digested_macs: list[DigestedMAC]):

    # Calculate the distances between the digested MACs in a distance matrix
    distance_matrix = calculate_distance_matrix(digested_macs)

    # Perform t-SNE
    tsne = TSNE(n_components=2, perplexity=5, n_iter=1000)
    tsne_results = tsne.fit_transform(distance_matrix)

    # Perform OPTICS clustering
    optics = OPTICS(min_samples=5, xi=0.1, min_cluster_size=5)
    optics.fit(tsne_results)

    # Get the labels of the clusters
    labels = optics.labels_

    if debug:

        # Plot the tne results with plotly
        import plotly.express as px
        import plotly.graph_objects as go
        # Convert the results to dicts
        tsne_results_dict = [{'x': tsne_results[i][0], 'y': tsne_results[i][1]} for i in range(len(tsne_results))]
        for i in range(len(digested_macs)):
            tsne_results_dict[i]['mac'] = digested_macs[i].mac
            tsne_results_dict[i]['signal (dbm)'] = int(10*log(digested_macs[i].average_signal_strenght/100000, 10)),
            tsne_results_dict[i]['presence_record'] = str(digested_macs[i].presence_record)
            tsne_results_dict[i]['ht_capabilities'] = digested_macs[i].ht_capabilities if digested_macs[i].ht_capabilities else ""
            tsne_results_dict[i]['ht_extended_capabilities'] = digested_macs[i].ht_extended_capabilities if digested_macs[i].ht_extended_capabilities else ""
            tsne_results_dict[i]['supported_rates'] = str(sorted(set(digested_macs[i].supported_rates))) if digested_macs[i].supported_rates else ""
            tsne_results_dict[i]['vendor_tags'] = str(sorted(set(digested_macs[i].tags))) if digested_macs[i].tags else ""
            tsne_results_dict[i]['optics_cluster'] = str(labels[i]) if labels[i] != -1 else "noise"
            # tsne_results_dict[i]['optics_cluster'] = labels[i]

        # Plot the t-SNE results with the digested_macs information
        fig = px.scatter(tsne_results_dict, x="x", y="y",
            # Dark dots
            color_discrete_sequence=["#e377c2", "#ff7f0e", "#2ca02c", "#d62728", "#9467bd", "#8c564b", "#1f77b4", "#7f7f7f", "#bcbd22", "#17becf"],
            color="optics_cluster",
            hover_data=["mac", "signal (dbm)", "presence_record",
                        "ht_capabilities", "ht_extended_capabilities", 
                        "supported_rates", "vendor_tags", "optics_cluster"],
            title="t-SNE and OPTICS results")


         # Increase the size of the dots
        fig.update_traces(marker=dict(size=5))

        # Set legend title to "Cluster"
        fig.update_layout(legend_title_text="Cluster")

        # Show the plot in pdf format plus current time in iso format
        fig.write("t-SNE_OPTICS_" + str(time.time()) + ".pdf")


        # fig.show()

    # Create a list of length max(labels)
    ret = [[] for i in range(max(labels) + 1)]
    noise = []


    # Create a list of tuples of the form (mac, cluster_id)
    for idx, l  in enumerate(labels):
        if l < 0:
            noise.append([digested_macs[idx].mac])
        else:
            ret[l].append(digested_macs[idx].mac)

    # Merge ret and noise
    ret = ret + noise
    
    return ret


# The distance matrix calculator
def calculate_distance_matrix(digested_macs: list[DigestedMAC]):
    # Create a distance matrix
    distance_matrix = [[0 for i in range(len(digested_macs))] for j in range(len(digested_macs))]

    # Calculate the distances between the digested MACs
    for i in range(len(digested_macs)):
        for j in range(len(digested_macs)):
            distance_matrix[i][j] = calculate_distance(digested_macs[i], digested_macs[j])

    if debug:
        # Show distance matrix in plotly
        import plotly.express as px
        # Plot the distance matrix
        fig = px.imshow(distance_matrix)
        # fig.show()
    
    return distance_matrix

# Definition of the calculator of the distance between two digested MACs
def calculate_distance(digested_mac_1: DigestedMAC, digested_mac_2: DigestedMAC):
    # Calculate the distance between the digested MACs
    distance = 0

    # Calculate the distance between the average signal strenght
    # The average signal strenght are converted to logarithmic scale
    # MAX 30
    dbm = abs(10*log(digested_mac_1.average_signal_strenght/100000, 10) - 10*log(digested_mac_2.average_signal_strenght/100000, 10))
    distance += normalize(dbm, 0, 100)*30


    # Calculate the distance between the HT capabilities
    # MAX 10
    if digested_mac_1.ht_capabilities is not None and digested_mac_2.ht_capabilities is not None:
        distance += 10 if digested_mac_1.ht_capabilities != digested_mac_2.ht_capabilities else 0
    else:
        distance += 2

    # Calculate the distance between the HT extended capabilities
    # MAX 10
    if digested_mac_1.ht_extended_capabilities is not None and digested_mac_2.ht_extended_capabilities is not None:
        distance += 10 if digested_mac_1.ht_extended_capabilities != digested_mac_2.ht_extended_capabilities else 0
    else:
        distance += 2

    # Calculate the distance between the supported rates
    # MAX 10
    if digested_mac_1.supported_rates != None and digested_mac_2.supported_rates != None:
        distance += 10 if set(digested_mac_1.supported_rates) != set(digested_mac_2.supported_rates) else 0
    else:
        distance += 2
    
    # Calculate the distance between the tags
    # MAX 10
    if digested_mac_1.tags != None and digested_mac_2.tags != None:
        distance += 10 if set(digested_mac_1.tags) != set(digested_mac_2.tags) else 0
    else:
        distance += 2
    
    ret = normalize(distance, 0, 70)

    return ret


# Function to normalize integer values to the range [0, 1]
def normalize(value: int, min_value: int, max_value: int):
    return (value - min_value) / (max_value - min_value)


# Run the fastAPI server
if __name__ == "__main__":
    import uvicorn
    uvicorn.run("clustering:app", host="0.0.0.0", port=8000, log_level="debug", reload=True)