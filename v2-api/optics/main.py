from fastapi import FastAPI
from numpy import ndarray
import uvicorn
from pydantic import BaseModel
from sklearn.cluster import OPTICS, cluster_optics_dbscan
import numpy as np

app = FastAPI()


class MacDigest(BaseModel):
    mac: str
    average_signal_strenght: float
    manufacturer: str | None
    oui_id: str | None
    type_count: list[int]
    presence_record: list[bool]
    ssid_probes: list[str] | None
    ht_capabilities: str | None
    ht_extended_capabilities: str | None
    supported_rates: list[str] | None
    tags: list[int] | None


# response_model te da el formato del return que prefieras
@app.post("/") #response_model=list(list(str)))
def optics(mac_digests: list[MacDigest]):
    digests = []
    for m in mac_digests:
        digests.append([
            m.average_signal_strenght,
            len(m.tags) if m.tags != None else 0
            ])

    clust = OPTICS(min_samples=2, max_eps=10, metric=d)
    clust.labels_ = [m.mac for m in mac_digests]
    result = clust.fit(digests)
    print("To return -> ", type(result.labels_))
    return {
        "result": result.labels_.tolist()
        }


def d(a, b):
    print("Called d!")
    print (a, b)
    return 1

def distance(m1, m2) -> int: 
    total = list()

    # Signal strength
    n = abs(m1.average_signal_strenght - m2.average_signal_strenght)/ max(m1.average_signal_strenght, m2.average_signal_strenght)
    total.append(n)

    # Manufacturer
    total.append(1 if m1.manufacturer == m2.manufacturer and m1.manufacturer != None else 0)

    # OUI
    total.append(1 if m1.oui_id == m2.oui_id and m1.oui_id != None else 0)

    # Type count difference
    n = 0
    m1_total = sum(m1.type_count)
    m2_total = sum(m2.type_count)
    for idx, m1t in enumerate(m1.type_count):
        n += m1t-m2.type_count[idx]
    total.append(n/max(m1_total, m2_total))

    # Presence record difference


    return sum(total)

if __name__ == "__main__":
    uvicorn.run("main:app", port=80, reload=True, host='0.0.0.0')
