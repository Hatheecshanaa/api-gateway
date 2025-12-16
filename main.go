package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/cors"
	"gopkg.in/yaml.v3"
)

// Config structs
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	JWTSecret string          `yaml:"jwt_secret"`
	Services  []ServiceConfig `yaml:"services"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type ServiceConfig struct {
	Name         string `yaml:"name"`
	PathPrefix   string `yaml:"path_prefix"`
	TargetURL    string `yaml:"target_url"`
	StripPrefix  string `yaml:"strip_prefix"`
	AuthRequired bool   `yaml:"auth_required"`
	EnvVar       string `yaml:"env_var"`
}

var logger *slog.Logger

// read config file and apply env overrides
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config yaml: %w", err)
	}

	// Environment overrides
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWTSecret = secret
	}

	for i := range cfg.Services {
		env := cfg.Services[i].EnvVar
		if env == "" {
			// default source for service URL
			n := strings.ToUpper(strings.ReplaceAll(cfg.Services[i].Name, "-", "_"))
			env = n + "_SERVICE_URL"
		}
		if v := os.Getenv(env); v != "" {
			cfg.Services[i].TargetURL = v
			logger.Info("service url overridden from env", "service", cfg.Services[i].Name, "var", env)
		}
	}

	return &cfg, nil
}

func newProxy(targetURL, stripPrefix string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target url: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	orig := proxy.Director
	proxy.Director = func(req *http.Request) {
		// keep user headers
		sub := req.Header.Get("X-User-Subject")
		userId := req.Header.Get("X-User-Id")
		roles := req.Header.Get("X-User-Roles")

		orig(req)
		req.Host = target.Host
		if sub != "" {
			req.Header.Set("X-User-Subject", sub)
		}
		if userId != "" {
			req.Header.Set("X-User-Id", userId)
		}
		if roles != "" {
			req.Header.Set("X-User-Roles", roles)
		}
		if stripPrefix != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, stripPrefix)
		}
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		logger.Info("response from downstream", "service", targetURL, "status", resp.Status, "path", resp.Request.URL.Path)
		return nil
	}

	return proxy, nil
}

// auth
type contextKey string

const userClaimsKey contextKey = "userClaims"

func authMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
				return
			}
			tok, found := strings.CutPrefix(auth, "Bearer ")
			if !found {
				http.Error(w, "Invalid Authorization Header format", http.StatusUnauthorized)
				return
			}
			p, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return secret, nil
			})
			if err != nil {
				logger.Warn("error parsing token", "err", err)
				http.Error(w, "Invalid Token", http.StatusUnauthorized)
				return
			}
			if claims, ok := p.Claims.(jwt.MapClaims); ok && p.Valid {
				ctx := context.WithValue(r.Context(), userClaimsKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
		})
	}
}

func injectUserInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := r.Context().Value(userClaimsKey).(jwt.MapClaims); ok {
			if sub, exists := claims["sub"]; exists {
				userIdStr := fmt.Sprintf("%v", sub)
				// Set both headers for compatibility with different services
				r.Header.Set("X-User-Subject", userIdStr)
				r.Header.Set("X-User-Id", userIdStr)
			}
			if roles, exists := claims["roles"]; exists {
				if rs, ok := roles.([]interface{}); ok {
					var parts []string
					for _, r := range rs {
						parts = append(parts, fmt.Sprintf("%v", r))
					}
					r.Header.Set("X-User-Roles", strings.Join(parts, ","))
				}
			}
			logger.Info("injecting user info headers", "sub", r.Header.Get("X-User-Subject"), "user-id", r.Header.Get("X-User-Id"))
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Command line flags
	cfgPath := flag.String("config", "config.yaml", "Path to configuration yaml")
	overridePort := flag.String("port", "", "Optional: override server port (e.g. :8080)")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Port override from flags
	if *overridePort != "" {
		cfg.Server.Port = *overridePort
	}

	r := buildRouter(cfg)

	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		logger.Info("api-gateway listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced shutdown", "err", err)
		os.Exit(1)
	}
	logger.Info("server exiting")
}

// buildRouter constructs a Chi router for the gateway — useful for testing
func buildRouter(cfg *Config) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	corsMw := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Subject", "X-User-Id", "X-User-Roles"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(corsMw.Handler)

	// health
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	authMw := authMiddleware([]byte(cfg.JWTSecret))

	for _, s := range cfg.Services {
		proxy, err := newProxy(s.TargetURL, s.StripPrefix)
		if err != nil {
			logger.Error("failed to create proxy", "service", s.Name, "err", err)
			os.Exit(1)
		}
		h := http.Handler(proxy)
		r.Group(func(r2 chi.Router) {
			if s.AuthRequired {
				r2.Use(authMw)
				r2.Use(injectUserInfo)
			}
			// Register both prefix and wildcard form to match both exact and nested paths
			r2.Handle(s.PathPrefix, h)
			r2.Handle(s.PathPrefix+"/*", h)
		})
		logger.Info("registered service", "name", s.Name, "prefix", s.PathPrefix, "target", s.TargetURL)
	}
	return r
}
