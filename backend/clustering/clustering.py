# t-SNE is used for dimensionality reduction
from sklearn.manifold import TSNE

# OPTICS is used for clustering
from sklearn.cluster import OPTICS

# Definition of a digested MAC in Pydantic model
from pydantic import BaseModel

# Import logarithms
from math import dist, log

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

@app.post("/cluster")
def receive_digested_macs(digested_macs: list[DigestedMAC]):

    # Calculate the distances between the digested MACs in a distance matrix
    distance_matrix = calculate_distance_matrix(digested_macs)

    # Perform t-SNE
    tsne = TSNE(n_components=2, perplexity=5, n_iter=300)
    tsne_results = tsne.fit_transform(distance_matrix)

    # Plot the tne results with plotly
    import plotly.express as px
    # Add the MAC addresses to the t-SNE results
    tsne_results2 = tsne_results.tolist()
    for i in range(len(digested_macs)):
        tsne_results2[i].append(digested_macs[i].mac)
    # Plot the t-SNE results
    fig = px.scatter(tsne_results2, x=0, y=1, color=2)
    fig.show()

    # Perform OPTICS clustering
    optics = OPTICS(min_samples=5, xi=0.05, min_cluster_size=0.05)
    optics.fit(tsne_results)

    # Get the labels of the clusters
    labels = optics.labels_

    # Show the results of optics clustering
    print("--------- results ----------")

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
    
    print(ret)

    return ret


# The distance matrix calculator
def calculate_distance_matrix(digested_macs: list[DigestedMAC]):
    # Create a distance matrix
    distance_matrix = [[0 for i in range(len(digested_macs))] for j in range(len(digested_macs))]

    # Calculate the distances between the digested MACs
    for i in range(len(digested_macs)):
        for j in range(len(digested_macs)):
            distance_matrix[i][j] = calculate_distance(digested_macs[i], digested_macs[j])

    # Show distance matrix in plotly
    import plotly.express as px
    # Plot the distance matrix
    fig = px.imshow(distance_matrix)
    # fig.show()
    
    return distance_matrix

# Definition of the calculator of the distance between two digested MACs
def calculate_distance(digested_mac_1: DigestedMAC, digested_mac_2: DigestedMAC):
    # Calculate the distance between the digested MACs
    distance = 0

    # Calculate the distance between the average signal strenght
    # The average signal strenght are converted to logarithmic scale
    dbm = abs(10*log(digested_mac_1.average_signal_strenght/100000, 10) - 10*log(digested_mac_2.average_signal_strenght/100000, 10))
    distance += normalize(dbm, 0, 100)*15
    # MAX 15


    # Calculate the distance between the manufacturer
    # if digested_mac_1.manufacturer is not None and digested_mac_2.manufacturer is not None:
    #     distance += 0 if digested_mac_1.manufacturer == digested_mac_2.manufacturer else 1
    
    # Calculate the distance between the OUI ID
    # MAX 2
    # distance += 2 if digested_mac_1.oui_id != digested_mac_2.oui_id else 0

    # Calculate the distance between the presence record
    # MAX 15
    # for i in range(len(digested_mac_1.presence_record)):
    #     if digested_mac_1.presence_record[i] or digested_mac_2.presence_record[i]:
    #         distance += 1 if digested_mac_1.presence_record[i] != digested_mac_2.presence_record[i] else 0

    # Calculate the distance between the HT capabilities
    # MAX 15
    if digested_mac_1.ht_capabilities is not None and digested_mac_2.ht_capabilities is not None:
        distance += 15 if digested_mac_1.ht_capabilities != digested_mac_2.ht_capabilities else 0
    else:
        distance += 5

    # Calculate the distance between the HT extended capabilities
    # MAX 15
    if digested_mac_1.ht_extended_capabilities is not None and digested_mac_2.ht_extended_capabilities is not None:
        distance += 15 if digested_mac_1.ht_extended_capabilities != digested_mac_2.ht_extended_capabilities else 0
    else:
        distance += 5

    # Calculate the distance between the supported rates
    # MAX 15
    if digested_mac_1.supported_rates != None and digested_mac_2.supported_rates != None:
        distance += 15 if set(digested_mac_1.supported_rates) != set(digested_mac_2.supported_rates) else 0
    else:
        distance += 5
    
    # Calculate the distance between the tags
    # MAX 15
    if digested_mac_1.tags != None and digested_mac_2.tags != None:
        distance += 15 if set(digested_mac_1.tags) != set(digested_mac_2.tags) else 0
    else:
        distance += 5
    
    ret = normalize(distance, 10, 75)
    print(digested_mac_1.mac, digested_mac_2.mac, ret)

    # if the macs are equal, print the distance
    if digested_mac_1.mac == digested_mac_2.mac:
        print("Equal:", digested_mac_1.mac, digested_mac_2.mac, distance)


    return ret


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    # Print request body 
    exc_str = f'{exc}'.replace('\n', ' ').replace('   ', ' ')
    logging.error(exc_str)
    content = {'status_code': 10422, 'message': exc_str, 'data': None}
    return JSONResponse(content=content, status_code=status.HTTP_422_UNPROCESSABLE_ENTITY)


# Function to normalize integer values to the range [0, 1]
def normalize(value: int, min_value: int, max_value: int):
    return (value - min_value) / (max_value - min_value)


# Run the fastAPI server
if __name__ == "__main__":
    import uvicorn
    uvicorn.run("clustering:app", host="0.0.0.0", port=8000, log_level="debug", reload=True)