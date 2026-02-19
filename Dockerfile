FROM golang:1.22-bullseye as deploy-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags "-w -s" -o app

# ---------------------------------------------------

FROM debian:bullseye-slim as deploy

RUN apt-get update

COPY --from=deploy-builder /app/app .

CMD ["./app"]

# ---------------------------------------------------
FROM golang:1.24-bullseye AS dev
WORKDIR /app
RUN go install github.com/air-verse/air@v1.61.7
CMD ["air"]