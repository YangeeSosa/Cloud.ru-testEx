package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"cloudru-balancer/internal/balancer"
	"cloudru-balancer/internal/config"
	"cloudru-balancer/internal/logger"
	"cloudru-balancer/internal/ratelimiter"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка чтения конфига: %v", err)
	}

	interval, err := time.ParseDuration(cfg.HealthCheck.Interval)
	if err != nil {
		log.Fatalf("Некорректный формат healthcheck.interval: %v", err)
	}

	b := balancer.New(cfg.Backends, interval, cfg.HealthCheck.Path)
	limiter := ratelimiter.NewRateLimiter(cfg.RateLimit.Capacity, cfg.RateLimit.Rate)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			logger.ErrorLogger.Printf("Не удалось определить IP клиента: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if !limiter.Allow(net.ParseIP(ip)) {
			logger.InfoLogger.Printf("Rate limit exceeded for %s", ip)
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"code":429,"message":"Rate limit exceeded"}`))
			return
		}
		b.ProxyRequest(w, r)
	})

	log.Printf("Балансировщик запущен на порту %s", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
