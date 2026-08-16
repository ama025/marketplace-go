package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"marketplace/internal/promotion/application/commands"
	"marketplace/internal/promotion/application/queries"
	promotiongrpc "marketplace/internal/promotion/grpc"
	"marketplace/internal/promotion/grpc/greetpb"
	"marketplace/internal/promotion/infrastructure/persistence"
)

func main() {

	for _, path := range []string{".env", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			log.Printf(".env loaded from %s", path)
			break
		}
	}
	log.Println("using environment variables")

	databaseURL := os.Getenv("PROMOTION_DATABASE_URL")
	grpcPort := os.Getenv("PROMOTION_GRPC_PORT")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("getwd error:", err)
	}
	migrationsPath := "file://" + cwd + "/migrations/promotion"

	if _, statErr := os.Stat(cwd + "/migrations/promotion"); statErr != nil {
		migrationsPath = "file://" + cwd + "/../../migrations/promotion"
	}
	log.Printf("migrations path: %s", migrationsPath)

	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		log.Fatal("mysql open error:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("mysql ping error:", err)
	}
	log.Println("mysql connected successfully")

	m, err := migrate.New(migrationsPath, "mysql://"+databaseURL)
	if err != nil {
		log.Fatal("migrations init error:", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("migrations up error:", err)
	}
	log.Println("migrations applied successfully")

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	repo := persistence.NewMySQLPromotionRepository(db)

	findOne := queries.NewFindByCatalogItemHandler(repo)
	findMany := queries.NewFindManyByCatalogItemsHandler(repo)

	addCmd := commands.NewAddDiscountHandler(repo)
	deactivateCmd := commands.NewDeactivateDiscountHandler(repo)

	grpcServer := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(grpcServer, promotiongrpc.NewGreetServer(findOne, findMany, addCmd, deactivateCmd))

	reflection.Register(grpcServer)

	log.Printf("promotion gRPC server started on :%s", grpcPort)

	metricsPort := os.Getenv("PROMOTION_METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "7013"
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("promotion metrics http server on :%s/metrics", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			log.Printf("metrics server error: %v", err)
		}
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("gRPC serve error:", err)
	}
}