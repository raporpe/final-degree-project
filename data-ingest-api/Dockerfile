FROM golang:1.18-alpine
WORKDIR /app


COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY cmd/ .

RUN go build -o app

CMD [ "./app" ]
