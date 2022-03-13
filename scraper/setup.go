package scraper

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"bitbucket.org/hilmarp/price-scraper/metrics"
	"github.com/gocolly/colly/v2"
)

func getCollector(allowedDomains []string, stackQueue string, stackParallel int, randomUserAgent bool) *colly.Collector {
	options := []colly.CollectorOption{
		colly.AllowedDomains(allowedDomains...),
		// colly.IgnoreRobotsTxt(),
	}

	if stackQueue == "stack" {
		options = append(options, colly.Async(true))
	}

	if !randomUserAgent {
		options = append(options, colly.UserAgent("Verdfra.is"))
	}

	c := colly.NewCollector(options...)

	limitRule := &colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 4 * time.Second,
	}

	if stackQueue == "stack" {
		limitRule.Parallelism = stackParallel
	}

	c.Limit(limitRule)

	c.SetRequestTimeout(30 * time.Second)

	if randomUserAgent {
		setRandomUserAgent(c)
	}

	return c
}

func setStorage(s Storage, collectors ...*colly.Collector) {
	for _, c := range collectors {
		err := c.SetStorage(s)
		if err != nil {
			log.Println(err)
		}
	}
}

func clearStorage(s Storage) {
	err := s.Clear()
	if err != nil {
		log.Println(err)
	}
}

func setEventHandlers(collectors ...*colly.Collector) {
	for _, c := range collectors {
		c.OnError(func(r *colly.Response, err error) {
			log.Println(fmt.Sprintf("Error scraping %s: %s", r.Request.URL.String(), err.Error()))
			metrics.ScraperErrorResponses.Inc()
		})

		c.OnResponse(func(r *colly.Response) {
			// fmt.Println("OnResponse", r.Request.URL)
			metrics.ScraperResponses.Inc()
		})
	}
}

var uaGens = []func() string{
	genFirefoxUA,
	genChromeUA,
}

// setRandomUserAgent generates a random browser user agent on every request
func setRandomUserAgent(c *colly.Collector) {
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", uaGens[rand.Intn(len(uaGens))]())
	})
}

var ffVersions = []float32{
	58.0,
	57.0,
	56.0,
	52.0,
	48.0,
	40.0,
	35.0,
}

var chromeVersions = []string{
	"65.0.3325.146",
	"64.0.3282.0",
	"41.0.2228.0",
	"40.0.2214.93",
	"37.0.2062.124",
}

var osStrings = []string{
	"Macintosh; Intel Mac OS X 10_10",
	"Windows NT 10.0",
	"Windows NT 5.1",
	"Windows NT 6.1; WOW64",
	"Windows NT 6.1; Win64; x64",
	"X11; Linux x86_64",
}

func genFirefoxUA() string {
	version := ffVersions[rand.Intn(len(ffVersions))]
	os := osStrings[rand.Intn(len(osStrings))]
	return fmt.Sprintf("Mozilla/5.0 (%s; rv:%.1f) Gecko/20100101 Firefox/%.1f", os, version, version)
}

func genChromeUA() string {
	version := chromeVersions[rand.Intn(len(chromeVersions))]
	os := osStrings[rand.Intn(len(osStrings))]
	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", os, version)
}
