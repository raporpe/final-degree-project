# Exploring the 802.11 protocol to estimate the degree of occupancy of an area
This repository hosts all my files for my final degree project at Carlos III University.

[Link to report](link.raporpe.me/final-degree-project)

## Summary

This project aims to estime the amount of people in an area leveraging 802.11 (WiFi) features and machine learning techniques.

The main idea is that WiFi devices (such as mobile phones) are constantly sending information when searching for APs (Access Points) to connect to. This data can be captured by simple devices with a WiFi antenna and then processed to extract information about the number of devices in the area since each device sends a unique identifier (MAC address) and most people have a mobile phone.


## Project structure


├───backend             # Golang program to receive, process, analyze and store data sent from capturing devices
│   ├───clustering          # Python API for performing t-SNE clustering on the data
│   └───cmd                 # Golang code for the backend
├───capture             # C++ program to capture data and send it to the backend (fast)
├───capture-python      # PoC to capture data with Python (was too slow for real-world use)
├───database            # SQL scripts to create the database where captured data is stored
├───experiments-data    # Raw data from real-world experiments that is explained in the report
├───frontend            # Simple web app for end-user visualization written in React
└───scripts             # Some Python scripts that were used during development for testing purposes


## System architecture

