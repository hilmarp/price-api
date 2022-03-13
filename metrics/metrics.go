package metrics

import "github.com/prometheus/client_golang/prometheus"

const namespace string = "verdfra"

var ScrapersRunning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "scrapers_running",
	Help:      "Number of scrapers currently running",
}, []string{"url"})

var ScrapersDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: namespace,
	Name:      "scrapers_duration",
	Help:      "How long for a scraper to finish",
}, []string{"url"})

var CleanersRunning = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "cleaners_running",
	Help:      "Number of cleaners currently running",
})

var CleanerDeleteCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "cleaner_delete_count",
	Help:      "Total products deleted by cleaner",
})

var WatchersRunning = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "watchers_running",
	Help:      "Number of watchers currently running",
})

var WatcherEmailsSent = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "watchers_emails_sent",
	Help:      "Number of watcher emails sent",
})

var WatcherSignups = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "watchers_signups",
	Help:      "Number of watcher signups",
})

var WatcherVerifies = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "watchers_verifies",
	Help:      "Number of watcher verifies",
})

var WatcherUnsubscribes = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "watchers_unsubscribes",
	Help:      "Number of watcher unsubscribes",
})

var TotalSearches = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "total_searches",
	Help:      "Number of searches",
})

var ViewCountersRunning = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "view_counters_running",
	Help:      "Number of view counters currently running",
})

var PriceChangeWatchersRunning = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "price_change_watchers_running",
	Help:      "Number of price change watchers currently running",
})

var ScraperResponses = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "scraper_responses",
	Help:      "Total scraper HTTP responses",
})

var ScraperErrorResponses = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "scraper_error_responses",
	Help:      "Total scraper error HTTP responses",
})

var ProductStoredCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "product_stored_count",
	Help:      "Total products stored",
})

var ProductStoredESCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "product_stored_es_count",
	Help:      "Total products stored in Elasticsearch",
})

var ScraperRedisEnqueues = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "scraper_redis_enqueues",
	Help:      "Scraper redis enqueues counter",
})

var ScraperRedisEnqueuesError = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "scraper_redis_enqueues_error",
	Help:      "Scraper redis enqueues error counter",
})

var ScraperRedisDequeues = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "scraper_redis_dequeues",
	Help:      "Scraper redis dequeues counter",
})

var ScraperRedisDequeuesError = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "scraper_redis_dequeues_error",
	Help:      "Scraper redis dequeues error counter",
})

var ProductClickCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "product_click_count",
	Help:      "Click on product URL",
})

// InitMetrics will register all metrics in the registry
func InitMetrics() {
	prometheus.MustRegister(ScrapersRunning)
	prometheus.MustRegister(ScrapersDuration)
	prometheus.MustRegister(CleanersRunning)
	prometheus.MustRegister(CleanerDeleteCount)
	prometheus.MustRegister(WatchersRunning)
	prometheus.MustRegister(WatcherEmailsSent)
	prometheus.MustRegister(WatcherSignups)
	prometheus.MustRegister(WatcherVerifies)
	prometheus.MustRegister(WatcherUnsubscribes)
	prometheus.MustRegister(TotalSearches)
	prometheus.MustRegister(ViewCountersRunning)
	prometheus.MustRegister(PriceChangeWatchersRunning)
	prometheus.MustRegister(ScraperResponses)
	prometheus.MustRegister(ScraperErrorResponses)
	prometheus.MustRegister(ProductStoredCount)
	prometheus.MustRegister(ProductStoredESCount)
	prometheus.MustRegister(ScraperRedisEnqueues)
	prometheus.MustRegister(ScraperRedisEnqueuesError)
	prometheus.MustRegister(ScraperRedisDequeues)
	prometheus.MustRegister(ScraperRedisDequeuesError)
	prometheus.MustRegister(ProductClickCount)
}
