FROM golang:1.21-alpine

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o rinha .

CMD ["/usr/src/app/rinha"]
