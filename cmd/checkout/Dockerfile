FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/checkout ./cmd/checkout

FROM alpine:3.20 AS final
WORKDIR /app
COPY --from=build /app/checkout .
COPY migrations/checkout ./migrations/checkout
EXPOSE 7004
ENTRYPOINT ["./checkout"]
