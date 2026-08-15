package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"os"

	_ "github.com/go-sql-driver/mysql"                      // MySQL драйвер для database/sql
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql" // драйвер MySQL для migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"    // источник миграций — папка с .sql файлами
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"marketplace/internal/promotion/application/commands"
	"marketplace/internal/promotion/application/queries"
	promotiongrpc "marketplace/internal/promotion/grpc"
	"marketplace/internal/promotion/grpc/greetpb"
	"marketplace/internal/promotion/infrastructure/persistence"
)

func main() {
	// Загружаем .env файл — пробуем несколько путей:
	// ".env"       — если запускаем из корня проекта (go run ./cmd/promotion/main.go)
	// "../../.env" — если запускаем из cmd/promotion/
	// В Docker переменные уже заданы через environment: в compose файле
	for _, path := range []string{".env", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			log.Printf(".env loaded from %s", path)
			break
		}
	}
	log.Println("using environment variables")

	// Читаем переменные окружения для подключения к MySQL
	// Формат DSN: user:pass@tcp(host:port)/dbname?parseTime=true&multiStatements=true
	// parseTime=true       — автоматически конвертирует DATETIME → time.Time
	// multiStatements=true — позволяет выполнять несколько SQL-запросов за раз (нужно для миграций)
	databaseURL := os.Getenv("PROMOTION_DATABASE_URL")
	grpcPort := os.Getenv("PROMOTION_GRPC_PORT")

	// Путь к миграциям строим динамически от текущей рабочей директории.
	// При запуске go run ./cmd/promotion/main.go из корня проекта — cwd = корень.
	// При запуске из cmd/promotion/ — поднимаемся на два уровня вверх.
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("getwd error:", err)
	}
	migrationsPath := "file://" + cwd + "/migrations/promotion"
	// Если файл не найден — пробуем подняться вверх (запуск из cmd/promotion/)
	if _, statErr := os.Stat(cwd + "/migrations/promotion"); statErr != nil {
		migrationsPath = "file://" + cwd + "/../../migrations/promotion"
	}
	log.Printf("migrations path: %s", migrationsPath)



	// --- Подключение к MySQL ---
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		log.Fatal("mysql open error:", err)
	}
	defer db.Close()

	// Проверяем соединение с базой данных
	if err := db.Ping(); err != nil {
		log.Fatal("mysql ping error:", err)
	}
	log.Println("mysql connected successfully")

	// --- Автомиграции ---
	// migrate.New принимает:
	//   1. Путь к папке с .sql файлами (из .env, формат: "file://путь/до/папки")
	//   2. DSN для подключения к MySQL
	m, err := migrate.New(migrationsPath, "mysql://"+databaseURL)
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

	// --- gRPC сервер ---
	// Слушаем TCP-порт из переменной окружения PROMOTION_GRPC_PORT
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	// --- Dependency Injection ---
	// Слой за слоем: DB → Repository → Handlers → gRPC Server
	repo := persistence.NewMySQLPromotionRepository(db)

	// Queries (read)
	findOne := queries.NewFindByCatalogItemHandler(repo)
	findMany := queries.NewFindManyByCatalogItemsHandler(repo)

	// Commands (write)
	addCmd := commands.NewAddDiscountHandler(repo)
	deactivateCmd := commands.NewDeactivateDiscountHandler(repo)

	// Создаём gRPC сервер и регистрируем наш GreetService
	grpcServer := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(grpcServer, promotiongrpc.NewGreetServer(findOne, findMany, addCmd, deactivateCmd))

	// Reflection позволяет Postman/grpcurl автоматически обнаруживать методы сервера
	reflection.Register(grpcServer)

	log.Printf("promotion gRPC server started on :%s", grpcPort)

	// Serve блокирует выполнение и обрабатывает входящие gRPC запросы
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("gRPC serve error:", err)
	}
}