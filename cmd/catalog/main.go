package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"marketplace/internal/catalog/api"
	"marketplace/internal/catalog/api/handlers"
	"marketplace/internal/catalog/application/commands"
	"marketplace/internal/catalog/application/queries"
	"marketplace/internal/catalog/infrastructure/persistence"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // драйвер postgres для migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // источник миграций — папка с .sql файлами
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

// main — точка входа в программу. Здесь инициализируется веб-сервер и маршруты.
func main() {

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println(err)
	}

	appPort := os.Getenv("CATALOG_APP_PORT")
	pgHost := os.Getenv("CATALOG_PG_HOST")
	pgPort := os.Getenv("CATALOG_PG_PORT")
	pgDB := os.Getenv("CATALOG_PG_DATABASE")
	pgUser := os.Getenv("CATALOG_PG_USER")
	pgPass := os.Getenv("CATALOG_PG_PASSWORD")
	pgSSL := os.Getenv("CATALOG_PG_SSLMODE")

	// CATALOG_MIGRATIONS_PATH задаётся в .env, например: file://../../migrations/catalog
	migrationsPath := os.Getenv("CATALOG_MIGRATIONS_PATH")

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

	brandRepo := persistence.NewBrandRepository(db)
	listBrandsHandler := queries.NewBrandsHandler(brandRepo)
	brandsHandler := handlers.NewBrandsHandler(listBrandsHandler)

	categoryRepo := persistence.NewCategoryRepository(db)
	listCategoriesHandler := queries.NewCategoriesHandler(categoryRepo)
	categoriesHandler := handlers.NewCategoriesHandler(listCategoriesHandler)

	itemRepo := persistence.NewItemRepository(db)
	listItemsHandler := queries.NewCatalogItemsHandler(itemRepo)
	itemByIDHandler := queries.NewCatalogItemByIDHandler(itemRepo)
	itemsByTitle := queries.NewCatalogItemByTitleHandler(itemRepo)
	itemsByBrand := queries.NewCatalogItemByBrandHandler(itemRepo) // Обработчик поиска товаров по бренду
	createItem := commands.NewCreateCatalogItemHandler(itemRepo)
	updateItem := commands.NewUpdateCatalogItemHandler(itemRepo)
	deleteItem := commands.NewDeleteCatalogItemHandler(itemRepo)
	itemsHandler := handlers.NewCatalogItemsHandler(listItemsHandler, itemByIDHandler, itemsByTitle, itemsByBrand, createItem, updateItem, deleteItem)

	r := gin.Default()

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	api.RegisterRoutes(r, brandsHandler, categoriesHandler, itemsHandler)

	if err := r.Run(":" + appPort); err != nil {
		log.Fatal(err)
	}
}
