package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	kafka "github.com/segmentio/kafka-go"
)

type ReadinessHandler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
	Kafka []string
}

func (h *ReadinessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{
		"postgres": checkDB(ctx, h.DB),
		"redis":    checkRedis(ctx, h.Redis),
		"kafka":    checkKafka(ctx, h.Kafka),
	}

	for _, status := range checks {
		if status != "ok" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "unhealthy", "checks": checks})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ready", "checks": checks})
}

func checkDB(ctx context.Context, pool *pgxpool.Pool) string {
	if err := pool.Ping(ctx); err != nil { return "failed" }
	return "ok"
}
func checkRedis(ctx context.Context, rdb *redis.Client) string {
	if err := rdb.Ping(ctx).Err(); err != nil { return "failed" }
	return "ok"
}
func checkKafka(ctx context.Context, brokers []string) string {
	conn, err := kafka.DialLeader(ctx, "tcp", brokers[0], "audit.logs.v1", 0)
	if err != nil { return "failed" }
	defer conn.Close()
	return "ok"
}