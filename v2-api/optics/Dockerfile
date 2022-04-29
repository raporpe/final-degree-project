FROM python:3.10

WORKDIR /app

COPY main.py .
COPY requirements.txt .


ENV FLASK_APP=/app/optics

RUN pip install --no-cache-dir --upgrade -r requirements.txt

CMD [ "python", "main.py"]
