FROM golang:1.26-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/${SERVICE_NAME}

FROM alpine:3.19
RUN apk add --no-cache ca-certificates netcat-openbsd

WORKDIR /app
COPY --from=build /out/service ./service
COPY certs ./certs

ENTRYPOINT ["./service"]
