package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"marketplace/internal/catalog/api"
	"marketplace/internal/catalog/api/handlers"
	"marketplace/internal/catalog/application/queries"
	"marketplace/internal/catalog/infrastructure/persistence"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

// main — точка входа в программу. Здесь инициализируется веб-сервер и маршруты.
func main() {

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}

	appPort := os.Getenv("CATALOG_APP_PORT")
	pgHost := os.Getenv("CATALOG_PG_HOST")
	pgPort := os.Getenv("CATALOG_PG_PORT")
	pgDB := os.Getenv("CATALOG_PG_DATABASE")
	pgUser := os.Getenv("CATALOG_PG_USER")
	pgPass := os.Getenv("CATALOG_PG_PASSWORD")
	pgSSL := os.Getenv("CATALOG_PG_SSLMODE")

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

	brandRepo := persistence.NewBrandRepository(db)
	listBrandsHandler := queries.NewBrandsHandler(brandRepo)
	brandsHandler := handlers.NewBrandsHandler(listBrandsHandler)

	r := gin.Default()

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	api.RegisterRoutes(r, brandsHandler)

	if err := r.Run(":" + appPort); err != nil {
		log.Fatal(err)
	}
}
