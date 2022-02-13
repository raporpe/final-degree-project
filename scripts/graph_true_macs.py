import psycopg2
import matplotlib.pyplot as plt
import numpy as np
import time


conn = psycopg2.connect("host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
cur = conn.cursor()

start_time = 1644505200
end_time = time.time()
t = start_time
graph = []
time = []
while t < end_time:
    cur.execute("SELECT COUNT(distinct station_mac) FROM probe_request_frames WHERE station_mac_vendor is not null and time < '{}'".format(t))
    number = cur.fetchall()[0][0]
    print(number)
    time.append(str(int((t-start_time)/3600/24)) + "d " + str((((t-start_time)/3600)+15)%24) + "h")
    graph.append(number)
    t += 3600

#plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%d-%h'))
#plt.gca().xaxis.set_major_locator(mdates.HourLocator(interval=5))

print(time)
print(graph)

plt.plot(time, graph)
ticks = np.arange(0, len(time)+1, 15)
plt.vlines(ticks, 0, 4000, color='red')
plt.xticks(ticks)
plt.savefig("./img.png") 
