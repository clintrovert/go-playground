FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR cmd/server

RUN go build -o server .

FROM alpine

COPY --from=0 /app/server /usr/bin

CMD ["/usr/bin/server"]