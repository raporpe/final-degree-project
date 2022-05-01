from dis import dis
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
            hash(m.manufacturer) if m.manufacturer != None else 0,
            hash(m.oui_id) if m.oui_id != None else 0,
            ])

    clust = OPTICS(min_samples=2, max_eps=10, metric=distance)
    clust.labels_ = [m.mac for m in mac_digests]
    result = clust.fit(digests)
    print("To return -> ", type(result.labels_))
    return result.labels_.tolist()



def distance(a, b) -> int: 
    total = list()

    # Signal strength - index 0
    n = abs(a[0] - b[0])/ max(a[0], b[0], 1)
    total.append(n)

    # Manufacturer - index 1
    total.append(1 if a[1] == b[1] and a[1] != None else 0)

    # OUI - index 2
    total.append(1 if a[2] == b[2] and a[2] != None else 0)

    # Type count difference - index 3
    #n = 0
    #m1_total = sum(a[3].type_count)
    #m2_total = sum(b[3].type_count)
    #for idx, m1t in enumerate(a[3].type_count):
    #    n += m1t-b[3].type_count[idx]
    #total.append(n/max(m1_total, m2_total))

    # Presence record difference


    return sum(total)

if __name__ == "__main__":
    uvicorn.run("main:app", port=80, reload=True, host='0.0.0.0')
