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
	"marketplace/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {

	for _, path := range []string{".env", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			log.Printf(".env loaded from %s", path)
			break
		}
	}

	appPort := os.Getenv("CATALOG_APP_PORT")
	pgHost := os.Getenv("CATALOG_PG_HOST")
	pgPort := os.Getenv("CATALOG_PG_PORT")
	pgDB := os.Getenv("CATALOG_PG_DATABASE")
	pgUser := os.Getenv("CATALOG_PG_USER")
	pgPass := os.Getenv("CATALOG_PG_PASSWORD")
	pgSSL := os.Getenv("CATALOG_PG_SSLMODE")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("getwd error:", err)
	}
	migrationsPath := fmt.Sprintf("file://%s/migrations/catalog", cwd)
	if _, statErr := os.Stat(cwd + "/migrations/catalog"); statErr != nil {
		migrationsPath = fmt.Sprintf("file://%s/../../migrations/catalog", cwd)
	}
	log.Printf("migrations path: %s", migrationsPath)

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
	itemsByBrand := queries.NewCatalogItemByBrandHandler(itemRepo)
	createItem := commands.NewCreateCatalogItemHandler(itemRepo)
	updateItem := commands.NewUpdateCatalogItemHandler(itemRepo)
	deleteItem := commands.NewDeleteCatalogItemHandler(itemRepo)
	itemsHandler := handlers.NewCatalogItemsHandler(listItemsHandler, itemByIDHandler, itemsByTitle, itemsByBrand, createItem, updateItem, deleteItem)

	r := gin.Default()
	r.Use(shared.MetricsMiddleware())

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/metrics", shared.MetricsHandler())

	api.RegisterRoutes(r, brandsHandler, categoriesHandler, itemsHandler)

	if err := r.Run(":" + appPort); err != nil {
		log.Fatal(err)
	}
}
