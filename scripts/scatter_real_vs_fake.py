import psycopg2
import plotly.express as px
import pandas as pd
import time
import datetime

conn = psycopg2.connect(
    "host=tfg-server.raporpe.dev dbname=tfg user=postgres password=raulportugues")
cur = conn.cursor()