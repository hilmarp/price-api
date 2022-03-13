package scraper

import (
	"log"
	"net/http"
	"time"

	"bitbucket.org/hilmarp/price-scraper/metrics"
	"github.com/robfig/cron/v3"
)

// StartCleaner will clean up non 200 status products in db/es
func (s *Scraper) StartCleaner() error {
	c := cron.New()
	c.AddFunc("0 0 * * 0", func() { // At 00:00 on Sunday
		metrics.CleanersRunning.Inc()
		defer metrics.CleanersRunning.Dec()

		limit := 100
		offset := 0

		for {
			products, err := s.DB.GetProducts(limit, offset, 0, 0, "id asc", "", []string{}, []string{})
			if err != nil {
				log.Print(err)
				break
			}

			if len(*products) == 0 {
				break
			}

			for _, product := range *products {
				req, err := http.NewRequest("GET", product.URL, nil)
				if err != nil {
					log.Print(err)
					continue
				}

				client := &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
					Timeout: 30 * time.Second,
				}

				// To not overload the website
				time.Sleep(500 * time.Millisecond)

				resp, err := client.Do(req)
				if err != nil {
					log.Print(err)
					continue
				}

				if resp.StatusCode != http.StatusOK {
					log.Printf("Cleaner is deleting product %s because %s", product.URL, resp.Status)
					err := s.DB.DeleteProductByID(product.ID)
					if err != nil {
						log.Print(err)
					}
					err = s.ES.DeleteSearchProductByID(product.ID)
					if err != nil {
						log.Print(err)
					}
					metrics.CleanerDeleteCount.Inc()
				}

				err = resp.Body.Close()
				if err != nil {
					log.Print(err)
				}
			}

			offset = offset + limit
		}
	})
	c.Start()

	return nil
}
