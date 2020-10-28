FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go build -o /go/src/app/simple-apm-start-allow-service

ENTRYPOINT ["/go/src/app/simple-apm-start-allow-service"]