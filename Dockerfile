# Build stage
FROM golang:1.20-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /apigateway

# Final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /apigateway /apigateway
COPY config.yaml /config.yaml
EXPOSE 8080
ENTRYPOINT ["/apigateway"]
CMD ["/apigateway", "-config", "/config.yaml"]
