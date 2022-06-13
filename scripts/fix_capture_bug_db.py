from dis import code_info
import psycopg2
import datetime as dt
import json
import os

# Get the database password from the environment variable
db_password = os.environ['DB_PASSWORD']

# Connect to the database
conn = psycopg2.connect("host=tfg-server.raporpe.dev dbname=tfg user=postgres password=" + db_password)
cur = conn.cursor()

# Start time from string
start_time = dt.datetime.fromisoformat("2022-03-26 15:10:00+00:00")
end_time = dt.datetime.fromisoformat("2022-03-27 15:10:00+00:00")

while start_time < end_time:
    
    # Get row from table detected_macs where start_time = start_time
    cur.execute("SELECT * FROM detected_macs WHERE start_time = %s", (start_time,))
    row = cur.fetchall()

    # If there is a row, get the detected_macs column and load json
    if len(row) == 1:
        data = row[0][4]
        j = json.loads(data)
        
        if j is None:
            continue

        # Traverse j if not null

        for key, value in j.items():
            if "signature" in value:
                continue
            else:
                print(start_time)
                exit()







    # Increase start_time by one minute
    start_time += dt.timedelta(minutes=1)
 