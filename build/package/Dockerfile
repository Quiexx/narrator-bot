FROM golang:1.21.5

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/**/*.go ./
COPY internal ./internal
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /server

ENTRYPOINT ["/server"]