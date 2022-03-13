package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"bitbucket.org/hilmarp/price-scraper/metrics"
	"github.com/go-redis/redis/v8"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

// Scraper handles scraping the web
type Scraper struct {
	DB              *SQL
	ES              *Elasticsearch
	Redis           *Redis
	Mongo           *Mongo
	Stop            chan struct{}
	RedisPort       string
	StackQueue      string
	QueueWorkers    int
	StackParallel   int
	Storage         string
	QueueStorage    string
	RandomUserAgent bool
}

type onlineStore struct {
	URL      string
	Selector string
	Callback colly.HTMLCallback
}

// StartScraper will start the web scraper
func (s *Scraper) StartScraper() error {
	var workers []func()
	onlineStores := s.getOnlineStores()
	redisDB := 10
	for _, os := range onlineStores {
		workers = append(workers, createScrapeWorker(os, s.StackQueue, s.Storage, s.QueueStorage, s.QueueWorkers, s.StackParallel, redisDB, s.RandomUserAgent, s.Mongo, s.DB))
		redisDB++
	}

	for _, worker := range workers {
		go func(worker func()) {
			for {
				worker()
			}
		}(worker)
	}

	<-s.Stop

	return nil
}

func createScrapeWorker(onlStore onlineStore, stackQueue, storageType, queueStorageType string, queueWorkers, stackParallel, redisDB int, randomUserAgent bool, mongo *Mongo, db *SQL) func() {
	return func() {
		startedAt := time.Now()
		env := os.Getenv("PRICE_APP_ENV")

		// Bot started and finished times
		bot, err := db.GetBotByURL(onlStore.URL)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Error getting bot from db: %s", err.Error())
				return
			}
		}

		clearStorageAtStart := true
		if env == "prod" && bot != nil {
			// If previous startedAt is before timeout we don't clear the storage, rather we continue where we left off last time
			diff := startedAt.Sub(bot.StartedAt)
			if diff < 12*time.Hour { // less than 12 hours ago
				log.Printf("Not clearing scraper %v storage because not enough time has passed", onlStore.URL)
				clearStorageAtStart = false
			}
		}

		// Metrics
		metrics.ScrapersRunning.WithLabelValues(onlStore.URL).Inc()
		prometheusTimer := prometheus.NewTimer(metrics.ScrapersDuration.WithLabelValues(onlStore.URL))

		// DB entry
		err = db.UpdateOrCreateBot(&Bot{URL: onlStore.URL, StartedAt: startedAt, FinishedAt: nil})
		if err != nil {
			log.Printf("Error updating/creating bot in db: %s", err.Error())
			return
		}

		// Set up mongo
		var mongoCli *Mongo = &Mongo{
			Client:     mongo.Client,
			Database:   mongo.Database,
			Collection: onlStore.URL,
		}

		// Set up redis
		var redisConn *redis.Client = redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:6379",
			DB:   redisDB,
		})
		redisCli := &Redis{Client: redisConn}

		// Storage to use
		var scraperStorage Storage = nil
		switch storageType {
		case "redis":
			scraperStorage = redisCli
		case "mongo":
			scraperStorage = mongoCli
		default:
			log.Println("Invalid scraper storage set")
			return
		}

		var scraperQueueStorage Storage = nil
		switch queueStorageType {
		case "redis":
			scraperQueueStorage = redisCli
		case "mongo":
			scraperQueueStorage = mongoCli
		default:
			log.Println("Invalid queue storage set")
			return
		}

		// Clear storage before
		if clearStorageAtStart {
			clearStorage(scraperStorage)
			clearStorage(scraperQueueStorage)
		}

		// Queue
		q, _ := queue.New(queueWorkers, scraperQueueStorage)

		hostURL := formatters.GetURLHost(onlStore.URL)

		allowedDomains := []string{
			hostURL, fmt.Sprintf("www.%s", hostURL),
		}

		c := getCollector(allowedDomains, stackQueue, stackParallel, randomUserAgent)

		setStorage(scraperStorage, c)
		setEventHandlers(c)

		// Set up the HTML matcher
		c.OnHTML(onlStore.Selector, onlStore.Callback)

		if hostURL == "byko.is" {
			c.OnResponse(func(r *colly.Response) {
				reqURL := r.Request.URL.String()
				if strings.Contains(reqURL, "?ProductID=") {
					return
				}

				if !strings.Contains(reqURL, "?feed=true") {
					url := fmt.Sprintf("%s?feed=true", reqURL)
					if stackQueue == "stack" {
						c.Visit(url)
					} else {
						q.AddURL(url)
					}
					return
				}

				type jsonResStruct struct {
					TotalPageCount int `json:"totalPageCount"`
					ProductList    []struct {
						ProdLink string `json:"prodLink"`
					} `json:"productList"`
				}
				var jsonRes jsonResStruct
				err := json.Unmarshal(r.Body, &jsonRes)
				if err == nil {
					for i := 2; i <= jsonRes.TotalPageCount; i++ {
						url := fmt.Sprintf("%s&PageNum=%v", reqURL, i)
						if stackQueue == "stack" {
							c.Visit(url)
						} else {
							q.AddURL(url)
						}
					}

					for _, product := range jsonRes.ProductList {
						url := strings.ReplaceAll(reqURL, "?feed=true", product.ProdLink)
						if stackQueue == "stack" {
							c.Visit(url)
						} else {
							q.AddURL(url)
						}
					}
				}
			})
		}

		// Visit all links with href attribute
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			url := e.Attr("href")
			url = e.Request.AbsoluteURL(url)
			if stackQueue == "stack" {
				c.Visit(url)
			} else {
				q.AddURL(url)
			}
		})

		if stackQueue == "stack" {
			c.Visit(onlStore.URL)
			c.Wait()
		} else {
			q.AddURL(onlStore.URL)
			q.Run(c)
		}

		// Cleanup
		metrics.ScrapersRunning.WithLabelValues(onlStore.URL).Dec()
		prometheusTimer.ObserveDuration()
		clearStorage(scraperStorage)
		clearStorage(scraperQueueStorage)
		redisConn.Close()

		// DB entry
		finishedAt := time.Now()
		err = db.UpdateOrCreateBot(&Bot{URL: onlStore.URL, StartedAt: startedAt, FinishedAt: &finishedAt})
		if err != nil {
			log.Printf("Error updating/creating bot in db: %s", err.Error())
		}
	}
}

// getOnlineStores returns online stores in priority order, index 0 has highest priority
func (s *Scraper) getOnlineStores() []onlineStore {
	var onlineStores []onlineStore

	env := os.Getenv("PRICE_APP_ENV")
	if env == "prod" {
		// PROD
		onlineStores = []onlineStore{
			{
				URL:      "https://elko.is/",
				Selector: "body.catalog-product-view",
				Callback: s.elkoCallback,
			},
			{
				URL:      "https://www.heimkaup.is/",
				Selector: ".ProductPage",
				Callback: s.heimkaupCallback,
			},
			{
				URL:      "https://rafha.is/",
				Selector: ".content-area.single-product",
				Callback: s.rafhaCallback,
			},
			{
				URL:      "https://ht.is/",
				Selector: "#product",
				Callback: s.htCallback,
			},
			{
				URL:      "https://www.rafland.is/",
				Selector: "#product",
				Callback: s.raflandCallback,
			},
			{
				URL:      "https://computer.is/",
				Selector: ".single-product",
				Callback: s.computerCallback,
			},
			{
				URL:      "https://ormsson.is/",
				Selector: ".product-details",
				Callback: s.ormssonCallback,
			},
			{
				URL:      "https://www.utilif.is/",
				Selector: "body.catalog-product-view",
				Callback: s.utilifCallback,
			},
			{
				URL:      "https://www.epal.is/",
				Selector: "body.single-product",
				Callback: s.epalCallback,
			},
			{
				URL:      "https://byko.is/",
				Selector: "#productListContentPlaceholder",
				Callback: s.bykoCallback,
			},
			{
				URL:      "https://tl.is/",
				Selector: ".product-head",
				Callback: s.tolvulistinnCallback,
			},
			{
				URL:      "https://nexus.is/",
				Selector: "body.single-product",
				Callback: s.nexusCallback,
			},
			{
				URL:      "https://www.rumfatalagerinn.is/",
				Selector: ".new-product-layout",
				Callback: s.rumfatalagerinnCallback,
			},
			{
				URL:      "https://www.penninn.is/",
				Selector: ".section__products",
				Callback: s.penninnCallback,
			},
		}
	} else {
		// DEV
		onlineStores = []onlineStore{
			{
				URL:      "https://elko.is/",
				Selector: "body.catalog-product-view",
				Callback: s.elkoCallback,
			},
			{
				URL:      "https://www.heimkaup.is/",
				Selector: ".ProductPage",
				Callback: s.heimkaupCallback,
			},
			{
				URL:      "https://rafha.is/",
				Selector: ".content-area.single-product",
				Callback: s.rafhaCallback,
			},
			{
				URL:      "https://ht.is/",
				Selector: "#product",
				Callback: s.htCallback,
			},
			{
				URL:      "https://www.rafland.is/",
				Selector: "#product",
				Callback: s.raflandCallback,
			},
			{
				URL:      "https://computer.is/",
				Selector: ".single-product",
				Callback: s.computerCallback,
			},
			{
				URL:      "https://ormsson.is/",
				Selector: ".product-details",
				Callback: s.ormssonCallback,
			},
			{
				URL:      "https://www.utilif.is/",
				Selector: "body.catalog-product-view",
				Callback: s.utilifCallback,
			},
			{
				URL:      "https://www.epal.is/",
				Selector: "body.single-product",
				Callback: s.epalCallback,
			},
			{
				URL:      "https://byko.is/",
				Selector: "#productListContentPlaceholder",
				Callback: s.bykoCallback,
			},
			{
				URL:      "https://tl.is/",
				Selector: ".product-head",
				Callback: s.tolvulistinnCallback,
			},
			{
				URL:      "https://nexus.is/",
				Selector: "body.single-product",
				Callback: s.nexusCallback,
			},
			{
				URL:      "https://www.rumfatalagerinn.is/",
				Selector: ".new-product-layout",
				Callback: s.rumfatalagerinnCallback,
			},
			{
				URL:      "https://www.penninn.is/",
				Selector: ".section__products",
				Callback: s.penninnCallback,
			},
		}
	}

	return onlineStores
}
