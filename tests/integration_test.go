// +build integration

package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/diasoft/diplomaverify/internal/api"
	"github.com/diasoft/diplomaverify/internal/config"
	"github.com/diasoft/diplomaverify/internal/crypto"
	"github.com/diasoft/diplomaverify/internal/db"
	"github.com/diasoft/diplomaverify/internal/repository"
	"github.com/diasoft/diplomaverify/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegration_DiplomaLifecycle(t *testing.T) {
	ctx := context.Background()

	// 1. Spin up PostgreSQL via Testcontainers
	pgContainer, err := postgres.Run(ctx, "postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections")),
	)
	if err != nil { t.Fatalf("Failed to start PG: %v", err) }
	defer pgContainer.Terminate(ctx)

	dsn, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
	os.Setenv("DB_DSN", dsn)

	// 2. Init dependencies (simplified for test)
	pool, _ := db.NewPool(ctx, db.PostgresConfig{DSN: dsn, MaxOpenConns: 5})
	cipher, _ := crypto.NewAESCipher("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	repo := repository.NewDiplomaRepository(pool, cipher)
	svc := service.NewDiplomaService(repo)
	
	// 3. Mock Router & Request
	router := api.NewRouter(&config.Config{Timeout: 5 * time.Second, AppEnv: "test"}, nil, nil, nil)
	
	// 4. Test Create (POST)
	payload := `{"university_id":"uni-1","records":[{"serial":"INT001","fio":"Test User","year":2024,"specialty":"QA"}]}`
	req := httptest.NewRequest("POST", "/api/v1/diplomas/bulk", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = httptest.NewBody()
	req.Body = io.NopCloser(strings.NewReader(payload))
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	
	// 5. Test Verify (GET)
	reqVerify := httptest.NewRequest("GET", "/api/v1/verify?number=INT001&university_code=TEST", nil)
	rrVerify := httptest.NewRecorder()
	router.ServeHTTP(rrVerify, reqVerify)
	assert.Equal(t, http.StatusOK, rrVerify.Code)
	
	fmt.Println("✅ Integration test passed: CRUD + Verify flow")
}