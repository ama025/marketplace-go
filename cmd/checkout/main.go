package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"marketplace/internal/checkout/api"
	"marketplace/internal/checkout/api/handlers"
	"marketplace/internal/checkout/application/commands"
	"marketplace/internal/checkout/application/queries"
	"marketplace/internal/checkout/infrastructure/persistence"
	"marketplace/internal/shared/messaging"
)

func main() {
	// Загружаем .env файл — пробуем несколько путей
	for _, path := range []string{".env", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			log.Printf(".env loaded from %s", path)
			break
		}
	}

	// --- Переменные окружения ---
	appPort     := os.Getenv("CHECKOUT_APP_PORT")
	databaseURL := os.Getenv("CHECKOUT_DATABASE_URL")
	rabbitHost  := os.Getenv("CHECKOUT_RABBITMQ_HOST")
	rabbitPort  := os.Getenv("CHECKOUT_RABBITMQ_PORT")
	rabbitUser  := os.Getenv("CHECKOUT_RABBITMQ_USER")
	rabbitPass  := os.Getenv("CHECKOUT_RABBITMQ_PASSWORD")

	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPass, rabbitHost, rabbitPort)

	// --- PostgreSQL ---
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("postgres open error:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("postgres ping error:", err)
	}
	log.Println("postgres connected successfully")

	// --- Автомиграции ---
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("getwd error:", err)
	}
	migrationsPath := fmt.Sprintf("file://%s/migrations/checkout", cwd)
	if _, statErr := os.Stat(cwd + "/migrations/checkout"); statErr != nil {
		migrationsPath = fmt.Sprintf("file://%s/../../migrations/checkout", cwd)
	}
	log.Printf("migrations path: %s", migrationsPath)

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		log.Fatal("migrations init error:", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("migrations up error:", err)
	}
	log.Println("migrations applied successfully")

	// --- Dependency Injection ---
	repo := persistence.NewPostgresOrderRepository(db)

	createOrderCmd := commands.NewCreateOrderHandler(repo)
	getByID        := queries.NewGetOrderByIDHandler(repo)
	getByAccount   := queries.NewGetOrdersByAccountHandler(repo)

	orderHandler := handlers.NewOrderHandler(createOrderCmd, getByID, getByAccount)

	// --- RabbitMQ Consumer ---
	// Checkout слушает события "OrderConfirmed" которые публикует Basket.
	// Запускаем в отдельной горутине — не блокирует HTTP сервер.
	consumer, err := messaging.NewConsumer(amqpURL)
	if err != nil {
		log.Fatal("rabbitmq consumer error:", err)
	}
	defer consumer.Close()

	// Контекст для graceful shutdown — останавливается по SIGTERM/SIGINT
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		// Handler вызывается для каждого входящего сообщения OrderConfirmed.
		// Создаёт заказ в PostgreSQL.
		err := consumer.Consume(ctx, func(ctx context.Context, event messaging.OrderConfirmedEvent) error {
			items := make([]commands.OrderItemInput, 0, len(event.Items))
			for _, i := range event.Items {
				id, err := uuid.Parse(i.ItemID)
				if err != nil {
					return fmt.Errorf("invalid item_id %s: %w", i.ItemID, err)
				}
				items = append(items, commands.OrderItemInput{
					ItemID:    id,
					ItemTitle: i.ItemTitle,
					Quantity:  i.Quantity,
					UnitPrice: i.UnitPrice,
					Discount:  i.Discount,
				})
			}

			order, err := createOrderCmd.Handle(ctx, commands.CreateOrderCommand{
				AccountName: event.AccountName,
				Items:       items,
			})
			if err != nil {
				return fmt.Errorf("create order: %w", err)
			}

			log.Printf("order created: id=%s account=%s total=%.2f",
				order.ID, order.AccountName, order.TotalPrice)
			return nil
		})

		if err != nil {
			log.Printf("rabbitmq consumer stopped: %v", err)
		}
	}()

	// --- HTTP сервер ---
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api.RegisterRoutes(r, orderHandler)

	log.Printf("checkout service started on :%s", appPort)
	if err := r.Run(":" + appPort); err != nil {
		log.Fatal(err)
	}
}
