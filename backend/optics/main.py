from dis import dis
from fastapi import FastAPI
from numpy import ndarray
import uvicorn
from pydantic import BaseModel
from sklearn.cluster import OPTICS, cluster_optics_dbscan
import numpy as np

app = FastAPI()

global_mem = {}

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
    for idx, m in enumerate(mac_digests):
        global_mem[idx] = m
        digests.append([
            hash(idx)
            ])

    clust = OPTICS(min_samples=2, max_eps=10, metric=distance)
    clust.labels_ = [m.mac for m in mac_digests]
    result = clust.fit(digests)
    print("To return -> ", type(result.labels_))
    return result.labels_.tolist()



def distance(a, b) -> int: 
    if type(a) == float or type(b) == float:
        print("-------------------------------")
    m1 : MacDigest = global_mem[int(a[0])]
    m2 : MacDigest = global_mem[int(b[0])]
    total = list()

    # Signal strength - index 0
    n = abs(m1.average_signal_strenght - m2.average_signal_strenght)/ max(m1.average_signal_strenght, m2.average_signal_strenght, 1)
    total.append(n)

    # Manufacturer - index 1
    total.append(1 if m1.manufacturer == m2.manufacturer and m1.manufacturer != None else 0)

    # OUI - index 2
    total.append(1 if m1.oui_id == m2.oui_id and m1.oui_id != None else 0)

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
