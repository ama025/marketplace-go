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
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // драйвер postgres для migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // источник миграций — папка с .sql файлами
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	basketAPI "marketplace/internal/basket/api"             // Маршруты basket-сервиса
	"marketplace/internal/basket/api/handlers"              // HTTP-хендлеры корзины
	"marketplace/internal/basket/infrastructure/cache"      // Реализация репозитория Redis
	"marketplace/internal/basket/infrastructure/persistence" // Реализация репозитория PostgreSQL
	"marketplace/internal/shared"                            // Общие middleware
	"marketplace/internal/shared/messaging"                  // RabbitMQ publisher
)

// main — точка входа в программу. Здесь инициализируется веб-сервер и маршруты.
func main() {

	// Загружаем .env файл (в Docker переменные уже заданы через environment:, поэтому ошибка не критична)
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

	// BASKET_MIGRATIONS_PATH задаётся в .env, например: file://../../migrations/basket
	migrationsPath := os.Getenv("BASKET_MIGRATIONS_PATH")

	// DSN для lib/pq (key=value формат)
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

	// --- Автомиграции ---
	// migrate.New принимает:
	//   1. Путь к папке с .sql файлами (из .env, формат: "file://путь/до/папки")
	//   2. DSN для подключения к PostgreSQL (формат URL для migrate отличается от lib/pq)
	migrateURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pgUser, pgPass, pgHost, pgPort, pgDB, pgSSL,
	)

	m, err := migrate.New(migrationsPath, migrateURL)
	if err != nil {
		log.Fatal("migrations init error:", err)
	}
	defer m.Close()

	// Up() применяет все новые (ещё не применённые) миграции.
	// ErrNoChange означает, что новых миграций нет — это норма, не ошибка.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("migrations up error:", err)
	}

	log.Println("migrations applied successfully")

	// --- Redis ---
	// Создаём клиент Redis, используя адрес и пароль из .env
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr, // например: localhost:9112
		Password: redisPass, // пароль из compose-dev.yaml
	})
	defer rdb.Close()

	// Проверяем соединение с Redis перед стартом сервера
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis ping error:", err)
	}
	log.Println("redis connected successfully")

	// --- Dependency Injection ---
	// Postgres-репозиторий остаётся — используется как fallback или для других операций
	_ = persistence.NewShoppingCartRepository(db)

	// Redis-репозиторий — основное хранилище корзины (быстрое, in-memory)
	cartRepo := cache.NewRedisCartRepository(rdb)

	// --- RabbitMQ Publisher ---
	// Basket публикует событие OrderConfirmed когда пользователь оформляет заказ.
	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPass, rabbitHost, rabbitPort)
	publisher, err := messaging.NewPublisher(amqpURL)
	if err != nil {
		// Не фатально — работаем без публикации (корзина продолжает работать)
		log.Printf("WARNING: rabbitmq publisher unavailable: %v", err)
		publisher = nil
	} else {
		defer publisher.Close()
		log.Println("rabbitmq publisher connected")
	}

	// Создаём HTTP-хендлер корзины, передавая интерфейс репозитория и publisher
	cartHandler := handlers.NewCartHandlerWithPublisher(cartRepo, publisher)

	// gin.New() — роутер без встроенных middleware (в отличие от gin.Default()).
	// Мы подключаем только наш ErrorHandleMiddleware, который покрывает и логирование ошибок, и recovery от паники.
	r := gin.New()
	r.Use(shared.ErrorHandleMiddleware())

	// Health-check эндпоинт для проверки работоспособности сервиса
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	// Регистрируем маршруты basket-сервиса
	basketAPI.RegisterRoutes(r, cartHandler)

	if err := r.Run(":" + appPort); err != nil {
		log.Fatal(err)
	}
}
