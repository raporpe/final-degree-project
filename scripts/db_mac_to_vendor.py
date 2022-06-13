from mac_vendor_lookup import MacLookup, VendorNotFoundError
import psycopg2
import os

MacLookup().update_vendors()


def vendor(mac):   
    try:
        return "'" + MacLookup().lookup(mac) + "'"
    except VendorNotFoundError as e:
        return None


# Get the database password from the environment variable
db_password = os.environ['DB_PASSWORD']

conn = psycopg2.connect("host=tfg-server.raporpe.dev dbname=tfg user=postgres password=" + db_password)
cur = conn.cursor()
cur.execute("SELECT * FROM probe_request ORDER BY time")
print(cur.rowcount)

rows = cur.fetchall()
for row in rows:
    print(row[6])
    if row[6] == None and True:
        vnd = vendor(row[2])
        if vnd is not None:
            print("Updating row at time {t} with vendor {v}".format(t=row[4], v=vnd))
            query = "UPDATE probe_request SET vendor = {vnd} WHERE id = '{id}';".format(vnd=vnd, id=row[0])
            print(query)
            cur.execute(query)
            conn.commit()


cur.close()




