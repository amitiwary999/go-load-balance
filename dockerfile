FROM golang:1.20-alpine

WORKDIR /server

COPY . .

RUN go mod download

RUN go build -o /load-balancer

EXPOSE 8080

CMD [ "/load-balancer" ]