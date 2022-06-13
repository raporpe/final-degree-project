# Using python 3.11
FROM python:3.10-bullseye

WORKDIR /app

# Install requirements
COPY requirements.txt .
RUN pip3 install -r requirements.txt

# Copy all the files in current directory
COPY . /app

# Run the app
CMD ["python", "clustering.py"]