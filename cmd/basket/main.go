package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	basketAPI "marketplace/internal/basket/api"
	"marketplace/internal/basket/api/handlers"
	"marketplace/internal/basket/infrastructure/cache"
	"marketplace/internal/basket/infrastructure/persistence"
	"marketplace/internal/shared"
	"marketplace/internal/shared/messaging"
)

func main() {

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println(err)
	}

	appPort    := os.Getenv("BASKET_APP_PORT")
	pgHost     := os.Getenv("BASKET_PG_HOST")
	pgPort     := os.Getenv("BASKET_PG_PORT")
	pgDB       := os.Getenv("BASKET_PG_DATABASE")
	pgUser     := os.Getenv("BASKET_PG_USER")
	pgPass     := os.Getenv("BASKET_PG_PASSWORD")
	pgSSL      := os.Getenv("BASKET_PG_SSLMODE")
	redisAddr  := os.Getenv("BASKET_REDIS_URL")
	redisPass  := os.Getenv("BASKET_REDIS_PASSWORD")
	rabbitHost := os.Getenv("BASKET_RABBITMQ_HOST")
	rabbitPort := os.Getenv("BASKET_RABBITMQ_PORT")
	rabbitUser := os.Getenv("BASKET_RABBITMQ_USER")
	rabbitPass := os.Getenv("BASKET_RABBITMQ_PASSWORD")

	migrationsPath := os.Getenv("BASKET_MIGRATIONS_PATH")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		pgHost,
		pgPort,
		pgUser,
		pgPass,
		pgDB,
		pgSSL,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("postgres ping error:", err)
	}

	migrateURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pgUser, pgPass, pgHost, pgPort, pgDB, pgSSL,
	)

	m, err := migrate.New(migrationsPath, migrateURL)
	if err != nil {
		log.Fatal("migrations init error:", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("migrations up error:", err)
	}

	log.Println("migrations applied successfully")

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
	})
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis ping error:", err)
	}
	log.Println("redis connected successfully")

	_ = persistence.NewShoppingCartRepository(db)

	cartRepo := cache.NewRedisCartRepository(rdb)

	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPass, rabbitHost, rabbitPort)
	publisher, err := messaging.NewPublisher(amqpURL)
	if err != nil {

		log.Printf("WARNING: rabbitmq publisher unavailable: %v", err)
		publisher = nil
	} else {
		defer publisher.Close()
		log.Println("rabbitmq publisher connected")
	}

	cartHandler := handlers.NewCartHandlerWithPublisher(cartRepo, publisher)

	r := gin.New()
	r.Use(shared.ErrorHandleMiddleware())
	r.Use(shared.MetricsMiddleware())

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/metrics", shared.MetricsHandler())

	basketAPI.RegisterRoutes(r, cartHandler)

	if err := r.Run(":" + appPort); err != nil {
		log.Fatal(err)
	}
}
