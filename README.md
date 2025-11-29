# API Gateway

This is a lightweight API Gateway implemented in Go using `chi` and `httputil.ReverseProxy`.
It reads service mappings from `config.yaml` and supports JWT authentication, header injection, CORS, and environment variable-based overrides.

Quickstart:

1. Build and run locally:

```bash
cd backend/api-gateway
# Build
GO111MODULE=on go build -o apigateway .
# Run
./apigateway
```

2. Override service URLs and JWT secret with environment variables when needed:

```bash
export JWT_SECRET="mysecret"
export PRODUCT_CATALOGUE_SERVICE_URL="http://localhost:8082"
export FRONTEND_ORIGINS="http://localhost:3000"
./apigateway
```

3. Use Docker:

```bash
cd backend/api-gateway
docker build -t cs02/apigateway:latest .
docker run -p 8080:8080 -e JWT_SECRET=mysecret cs02/apigateway:latest
```

4. Run unit tests:

```bash
go test ./...
```

Notes:
- By default the gateway listens on :8080, adjust `config.yaml:server.port` as necessary.
- This gateway uses HS256 tokens for JWT parsing; if you use RSA tokens (RS256), the parsing logic must be updated.
