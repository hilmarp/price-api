package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"bitbucket.org/hilmarp/price-scraper/metrics"
	"bitbucket.org/hilmarp/price-scraper/scraper"
	"bitbucket.org/hilmarp/price-scraper/web"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	exPath := filepath.Dir(ex)
	err = godotenv.Load(exPath + "/../.env")
	if err != nil {
		log.Fatal(err)
	}

	// Context
	ctx := context.Background()

	// Rand seed
	rand.Seed(time.Now().UnixNano())

	// Metrics
	metrics.InitMetrics()

	// Init MySQL
	dbUser := os.Getenv("PRICE_SQL_USER")
	dbPass := os.Getenv("PRICE_SQL_PASSWORD")
	dbDb := os.Getenv("PRICE_SQL_DB")
	dbPort := os.Getenv("PRICE_SQL_PORT")
	connStr := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbPort, dbDb)

	scraperDB, err := gorm.Open(mysql.Open(connStr), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}

	webDB, err := gorm.Open(mysql.Open(connStr), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Init Elasticsearch
	scraperES, err := elastic.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	webES, err := elastic.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Init redis
	redisPort := os.Getenv("PRICE_REDIS_PORT")

	webRedis := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("127.0.0.1:%s", redisPort),
		DB:   1,
	})
	_, err = webRedis.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}
	defer webRedis.Close()

	// Init MongoDB
	scraperMongo, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer scraperMongo.Disconnect(ctx)

	// Handle CTRL^C
	sigChan := make(chan os.Signal, 1)
	stopChan := make(chan struct{}, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		stopChan <- struct{}{}
		os.Exit(0)
	}()

	// Start scraper
	scrapeStackQueue := os.Getenv("PRICE_STACK_QUEUE")
	scrapeStorage := os.Getenv("PRICE_SCRAPE_STORAGE")
	scrapeQueueStorage := os.Getenv("PRICE_SCRAPE_QUEUE_STORAGE")
	scrapeQueueWorkersStr := os.Getenv("PRICE_QUEUE_WORKERS")
	scrapeStackParallelStr := os.Getenv("PRICE_STACK_PARALLEL")
	scrapeRandomUserAgentStr := os.Getenv("PRICE_RANDOM_USER_AGENT")
	scrapeQueueWorkers := 2
	scrapeStackParallel := 2
	scrapeRandomUserAgent := false

	num, err := strconv.Atoi(scrapeQueueWorkersStr)
	if err == nil {
		scrapeQueueWorkers = num
	}

	num, err = strconv.Atoi(scrapeStackParallelStr)
	if err == nil {
		scrapeStackParallel = num
	}

	if scrapeStackQueue == "" {
		scrapeStackQueue = "stack"
	}

	if scrapeStorage == "" {
		scrapeStorage = "redis"
	}

	if scrapeQueueStorage == "" {
		scrapeQueueStorage = "redis"
	}

	if scrapeRandomUserAgentStr == "true" {
		scrapeRandomUserAgent = true
	}

	scraperDBInit := &scraper.SQL{DB: scraperDB}
	scraperService := scraper.Scraper{
		DB:              scraperDBInit,
		ES:              &scraper.Elasticsearch{Client: scraperES},
		Redis:           nil,
		Mongo:           &scraper.Mongo{Client: scraperMongo, Database: "price"},
		Stop:            stopChan,
		RedisPort:       redisPort,
		StackQueue:      scrapeStackQueue,
		QueueWorkers:    scrapeQueueWorkers,
		StackParallel:   scrapeStackParallel,
		Storage:         scrapeStorage,
		QueueStorage:    scrapeQueueStorage,
		RandomUserAgent: scrapeRandomUserAgent,
	}

	// Run db migration
	err = scraperDBInit.Migrate()
	if err != nil {
		log.Fatal(err)
	}

	go scraperService.StartScraper()
	go scraperService.StartCleaner()
	go scraperService.StartWatcher()
	go scraperService.StartViewCounter()
	go scraperService.StartPriceChangeWatcher()

	// Start API server
	apiServer := web.APIServer{
		DB:    &scraper.SQL{DB: webDB},
		ES:    &scraper.Elasticsearch{Client: webES},
		Redis: &web.Redis{Client: webRedis},
		Port:  os.Getenv("PRICE_WEB_SERVER_PORT"),
	}
	err = apiServer.StartServer()
	if err != nil {
		log.Fatal(err)
	}
}
