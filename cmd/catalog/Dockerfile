FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/catalog ./cmd/catalog

FROM alpine:3.20 AS final
WORKDIR /app
COPY --from=build /app/catalog .
COPY migrations/catalog ./migrations/catalog
EXPOSE 7001
ENTRYPOINT ["./catalog"]