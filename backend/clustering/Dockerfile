# Using python 3.10
FROM python:3.10-alpine

# Install dependencies
RUN apk add --no-cache python3-pip

# Copy all the files in current directory
COPY . /app

# Install requirements
RUN pip3 install -r requirements.txt

# Run the app
WORKDIR /app
CMD ["python", "clustering.py"]