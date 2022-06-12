# t-SNE is used for dimensionality reduction
from sklearn.manifold import TSNE

# OPTICS is used for clustering
from sklearn.cluster import OPTICS

# Definition of a digested MAC in Pydantic model
from pydantic import BaseModel

# Import types for defining the model
from typing import List, Any

# Import logarithms
from math import log

# The model of a digested MAC to cluster
class DigestedMAC(BaseModel):
    mac: str
    average_signal_strenght: float
    manufacturer: str | None
    oui_id: str
    type_count: List[int]
    presence_record: List[bool]
    ssid_probes: List[str]
    ht_capabilities: str | None
    ht_extended_capabilities: Any | None
    supported_rates: List[float]
    tags: List[int]


# FastAPI will be used for receiving data from Go
from fastapi import FastAPI

# Define the single API route for receiving digested MACs
app = FastAPI()

@app.post("/digested-macs")
def receive_digested_macs(digested_macs: List[DigestedMAC]):
    # Perform dimensionality reduction with t-SNE

    # Calculate the distances between the digested MACs in a distance matrix
    distance_matrix = calculate_distance_matrix(digested_macs)

    # Perform t-SNE
    tsne = TSNE(n_components=2, perplexity=30, n_iter=300)
    tsne_results = tsne.fit_transform(distance_matrix)

    # Perform OPTICS clustering
    optics = OPTICS(min_samples=5, xi=0.05, min_cluster_size=0.05)
    optics.fit(tsne_results)

    # Get the labels of the clusters
    labels = optics.labels_

    # Show the results of optics clustering
    print(labels)

    # Return the labels of the clusters
    return labels


# The distance matrix calculator
def calculate_distance_matrix(digested_macs: List[DigestedMAC]):
    # Create a distance matrix
    distance_matrix = [[0 for i in range(len(digested_macs))] for j in range(len(digested_macs))]

    # Calculate the distances between the digested MACs
    for i in range(len(digested_macs)):
        for j in range(len(digested_macs)):
            distance_matrix[i][j] = calculate_distance(digested_macs[i], digested_macs[j])
    
    return distance_matrix

# Definition of the calculator of the distance between two digested MACs
def calculate_distance(digested_mac_1: DigestedMAC, digested_mac_2: DigestedMAC):
    # Calculate the distance between the digested MACs
    distance = 0

    # Calculate the distance between the average signal strenght
    # The average signal strenght are converted to logarithmic scale
    distance += abs(log(digested_mac_1.average_signal_strenght) - log(digested_mac_2.average_signal_strenght))

    # Calculate the distance between the manufacturer
    if digested_mac_1.manufacturer is not None and digested_mac_2.manufacturer is not None:
        distance += 0 if digested_mac_1.manufacturer == digested_mac_2.manufacturer else 1
    
    # Calculate the distance between the OUI ID
    distance += 0 if digested_mac_1.oui_id == digested_mac_2.oui_id else q

    # Calculate the distance between the type count

    # Calculate the distance between the presence record
    for i in range(len(digested_mac_1.presence_record)):
        distance += 1 if digested_mac_1.presence_record[i] != digested_mac_2.presence_record[i] else 0

    # Calculate the distance between the SSID probes.

    # Calculate the distance between the HT capabilities
    if digested_mac_1.ht_capabilities is not None and digested_mac_2.ht_capabilities is not None:
        distance += 0 if digested_mac_1.ht_capabilities == digested_mac_2.ht_capabilities else 1

    # Calculate the distance between the HT extended capabilities
    if digested_mac_1.ht_extended_capabilities is not None and digested_mac_2.ht_extended_capabilities is not None:
        distance += 0 if digested_mac_1.ht_extended_capabilities == digested_mac_2.ht_extended_capabilities else 1

    # Calculate the distance between the supported rates
    if digested_mac_1.supported_rates != digested_mac_2.supported_rates:
        distance += 1
    
    # Calculate the distance between the tags
    if digested_mac_1.tags != digested_mac_2.tags:
        distance += 1
    
    return distance
