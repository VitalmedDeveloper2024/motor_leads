FROM golang:1.22 AS build
WORKDIR /src
COPY . .
# Usa as dependências em ./vendor (geradas no host com 'go mod vendor').
# Build totalmente offline — não acessa proxy.golang.org.
RUN CGO_ENABLED=0 go build -mod=vendor -o /bin/motor-leads ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/motor-leads /app/motor-leads
COPY frontend ./frontend
EXPOSE 8080
ENTRYPOINT ["/app/motor-leads"]
