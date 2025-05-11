package balancer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"cloudru-balancer/internal/logger"
)

// Balancer описывает балансировщик нагрузки
type Balancer struct {
	backends      []string
	alive         map[string]bool
	current       int
	mutex         sync.Mutex
	checkInterval time.Duration
	healthPath    string
}

// New создает новый балансировщик с заданными бэкендами и параметрами health-check
func New(backends []string, checkInterval time.Duration, healthPath string) *Balancer {
	b := &Balancer{
		backends:      backends,
		alive:         make(map[string]bool),
		current:       0,
		checkInterval: checkInterval,
		healthPath:    healthPath,
	}
	for _, backend := range backends {
		b.alive[backend] = true
	}
	go b.healthCheckLoop()
	return b
}

// healthCheckLoop периодически проверяет доступность бэкендов
func (b *Balancer) healthCheckLoop() {
	for {
		b.checkBackends()
		time.Sleep(b.checkInterval)
	}
}

func (b *Balancer) checkBackends() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, backend := range b.backends {
		resp, err := http.Get(backend + b.healthPath)
		if err != nil || resp.StatusCode != 200 {
			b.alive[backend] = false
			logger.ErrorLogger.Printf("Backend %s недоступен", backend)
			continue
		}
		b.alive[backend] = true
		resp.Body.Close()
	}
}

// NextBackend выбирает следующий доступный бэкенд по алгоритму round-robin
func (b *Balancer) NextBackend() (string, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for i := 0; i < len(b.backends); i++ {
		idx := (b.current + i) % len(b.backends)
		backend := b.backends[idx]
		if b.alive[backend] {
			b.current = (idx + 1) % len(b.backends)
			return backend, nil
		}
	}
	return "", http.ErrServerClosed // нет доступных бэкендов
}

// ProxyRequest проксирует запрос на выбранный бэкенд
func (b *Balancer) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	backend, err := b.NextBackend()
	if err != nil {
		logger.ErrorLogger.Printf("Нет доступных бэкендов")
		http.Error(w, "No available backends", http.StatusServiceUnavailable)
		return
	}
	logger.InfoLogger.Printf("%s %s -> %s", r.Method, r.URL.Path, backend)
	backendURL, err := url.Parse(backend)
	if err != nil {
		logger.ErrorLogger.Printf("Ошибка парсинга backend URL: %v", err)
		http.Error(w, "Bad backend URL", http.StatusInternalServerError)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
		logger.ErrorLogger.Printf("Ошибка проксирования запроса: %v", e)
		http.Error(rw, "Backend unavailable", http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, r)
}
