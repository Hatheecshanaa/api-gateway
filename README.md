# 🚪 API Gateway

> Central entry point for all API requests in the CS02 E-Commerce Platform

## 📋 Overview

This is a lightweight API Gateway implemented in Go using `chi` router and `httputil.ReverseProxy`. It serves as the single entry point for all frontend requests, handling JWT authentication, request routing, CORS, and header injection.

## 🛠️ Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.20+ |
| Router | Chi (go-chi/chi) | v5 |
| JWT | golang-jwt | v4 |
| CORS | go-chi/cors | Latest |
| Config | gopkg.in/yaml.v3 | Latest |

## 🚀 Route Configuration

| Path Prefix | Target Service | Port | Auth Required |
|-------------|----------------|------|---------------|
| `/api/auth/*` | user-identity-service | 8081 | No |
| `/api/users/*` | user-identity-service | 8081 | Yes |
| `/api/products/*` | product-catalogue-service | 8082 | No |
| `/api/builder/*` | product-catalogue-service | 8082 | No |
| `/api/cart/*` | shoppingcart-wishlist-service | 8084 | Yes |
| `/api/wishlist/*` | shoppingcart-wishlist-service | 8084 | Yes |
| `/api/orders/*` | order-service | 8083 | Yes |
| `/api/trade-in/*` | order-service | 8083 | Yes |
| `/api/notifications/*` | notifications-service | 8087 | Yes |
| `/api/content/*` | content-service | 8086 | No |
| `/api/support/*` | support-service | 8085 | Yes |
| `/api/warranty/*` | support-service | 8085 | Yes |
| `/api/analytics/*` | reporting-and-analysis-service | 8088 | Yes |
| `/api/ai/*` | AI-service | 8089 | No |
| `/healthz` | Gateway health check | - | No |

## 🔧 Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes | - | Secret key for JWT validation |
| `FRONTEND_ORIGINS` | No | `http://localhost:3000` | Allowed CORS origins |
| `USER_IDENTITY_SERVICE_URL` | No | `http://localhost:8081` | User service URL |
| `PRODUCT_CATALOGUE_SERVICE_URL` | No | `http://localhost:8082` | Product service URL |
| `ORDER_SERVICE_URL` | No | `http://localhost:8083` | Order service URL |
| `CART_SERVICE_URL` | No | `http://localhost:8084` | Cart service URL |
| `SUPPORT_SERVICE_URL` | No | `http://localhost:8085` | Support service URL |
| `CONTENT_SERVICE_URL` | No | `http://localhost:8086` | Content service URL |
| `NOTIFICATIONS_SERVICE_URL` | No | `http://localhost:8087` | Notifications service URL |
| `ANALYTICS_SERVICE_URL` | No | `http://localhost:8088` | Analytics service URL |
| `AI_SERVICE_URL` | No | `http://localhost:8089` | AI service URL |

### config.yaml

```yaml
server:
  port: 8080

services:
  user-identity-service:
    url: http://localhost:8081
  product-catalogue-service:
    url: http://localhost:8082
  # ... additional services

jwt:
  secret: ${JWT_SECRET}

cors:
  allowed_origins:
    - http://localhost:3000
```

## 📦 Dependencies

```go
require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/go-chi/cors v1.2.1
    github.com/golang-jwt/jwt/v4 v4.5.0
    gopkg.in/yaml.v3 v3.0.1
)
```

## 🏃 Running the Service

### Local Development

```bash
cd backend/api-gateway

# Build
GO111MODULE=on go build -o apigateway .

# Run
./apigateway
```

### With Environment Variables

```bash
export JWT_SECRET="mysecret"
export PRODUCT_CATALOGUE_SERVICE_URL="http://localhost:8082"
export FRONTEND_ORIGINS="http://localhost:3000"
./apigateway
```

### Docker

```bash
cd backend/api-gateway

# Build image
docker build -t cs02/apigateway:latest .

# Run container
docker run -p 8080:8080 \
  -e JWT_SECRET=mysecret \
  -e USER_IDENTITY_SERVICE_URL=http://user-identity-service:8081 \
  cs02/apigateway:latest
```

### Using Makefile

```bash
make build    # Build the binary
make run      # Run the gateway
make test     # Run tests
make docker   # Build Docker image
```

## 🔐 Authentication Flow

1. **Public Routes**: Requests to `/api/auth/*`, `/api/products/*`, `/api/content/*`, `/api/ai/*` pass through without auth
2. **Protected Routes**: All other routes require a valid JWT token
3. **Token Extraction**: JWT is extracted from `Authorization: Bearer <token>` header
4. **Validation**: Token signature verified using HS256 algorithm
5. **Header Injection**: On successful auth, gateway injects:
   - `X-User-Subject`: User's subject claim
   - `X-User-Id`: User's ID claim
   - `X-User-Roles`: User's roles claim

## ✅ Features - Completion Status

| Feature | Status | Notes |
|---------|--------|-------|
| Reverse proxy routing | ✅ Complete | All services routed |
| JWT authentication middleware | ✅ Complete | HS256 tokens |
| User info header injection | ✅ Complete | X-User-Id, X-User-Subject, X-User-Roles |
| CORS handling | ✅ Complete | Configurable origins |
| Health check endpoint | ✅ Complete | `/healthz` |
| Environment variable config | ✅ Complete | Override via env vars |
| YAML configuration | ✅ Complete | `config.yaml` |
| Graceful shutdown | ✅ Complete | Handles SIGTERM |
| Request logging | ✅ Complete | Chi middleware |

### **Overall Completion: 100%** ✅

## ❌ Not Implemented / Future Enhancements

| Feature | Priority | Notes |
|---------|----------|-------|
| Rate limiting | Medium | Recommended for production |
| Request caching | Low | Could cache product requests |
| Circuit breaker | Medium | For service resilience |
| API versioning | Low | Currently v1 only |
| RS256 JWT support | Low | Currently HS256 only |
| Metrics/Prometheus | Medium | For monitoring |
| Distributed tracing | Low | OpenTelemetry integration |

## 📁 Project Structure

```
api-gateway/
├── main.go          # Application entry point
├── main_test.go     # Unit tests
├── config.yaml      # Service configuration
├── go.mod           # Go module definition
├── Dockerfile       # Container configuration
├── Makefile         # Build commands
├── run.sh           # Development script
└── README.md        # This file
```

## 🧪 Testing

```bash
# Run unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Test health endpoint
curl http://localhost:8080/healthz

# Test with JWT token
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/users/me
```

## 🔗 Related Services

All backend microservices are routed through this gateway:
- [User Identity Service](../user-identity-service/README.md)
- [Product Catalogue Service](../product-catalogue-service/README.md)
- [Order Service](../order-service/README.md)
- [Shopping Cart Service](../shoppingcart-wishlist-service/README.md)
- [Support Service](../support-service/README.md)
- [Content Service](../content-service/README.md)
- [Notifications Service](../notifications-service/README.md)
- [Reporting Service](../reporting-and-analysis-service/README.md)
- [AI Service](../AI-service/README.md)

## 📝 Notes

- Default port is **8080**, configurable via `config.yaml`
- Uses **HS256** tokens; update parsing logic for RS256
- All service URLs can be overridden via environment variables
- CORS is configured to allow frontend origins
