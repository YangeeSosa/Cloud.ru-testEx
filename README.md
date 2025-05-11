# Cloud.ru-balancer

Простой HTTP-балансировщик нагрузки.

## Возможности
- Балансировка входящих HTTP-запросов по пулу бэкендов (round-robin)
- Проверка доступности бэкендов (health-check)
- Ограничение частоты запросов для клиентов (rate-limiting)
- Гибкая настройка через YAML-конфиг
- Логирование запросов и ошибок

## Быстрый старт

### 1. Клонируйте репозиторий и установите зависимости
```sh
go mod tidy
```

### 2. Настройте конфиг `configs/config.yaml`
```yaml
port: ":8080"
backends:
  - "http://localhost:8081"
  - "http://localhost:8082"
ratelimit:
  capacity: 10
  rate: 5
healthcheck:
  interval: 3s
  path: "/health"
```
- `port` — порт балансировщика
- `backends` — список адресов бэкендов
- `ratelimit.capacity` — максимальное количество токенов (burst)
- `ratelimit.rate` — скорость пополнения токенов (в секунду)
- `healthcheck.interval` — интервал проверки бэкендов
- `healthcheck.path` — путь для health-check (должен возвращать 200)

### 3. Запустите балансировщик
```sh
go run cmd/main.go
```

### 4. Запустите бэкенды (пример на Go):
```go
package main
import (
	"fmt"
	"net/http"
)
func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Backend response")
	})
	http.ListenAndServe(":8081", nil) // или :8082 для второго бэкенда
}
```

### 5. Проверьте работу
- Откройте http://localhost:8080 — запросы будут распределяться по бэкендам.
- Если лимит превышен — получите 429.
- Если все бэкенды недоступны — получите 503.

### 6. Тестирование нагрузки
```sh
ab -n 500 -c 100 http://localhost:8080/
```

## Структура проекта
- `cmd/main.go` — точка входа
- `internal/balancer` — балансировка и health-check
- `internal/ratelimiter` — rate-limiting
- `internal/config` — работа с конфигом
- `internal/logger` — логирование
- `configs/config.yaml` — пример конфига
