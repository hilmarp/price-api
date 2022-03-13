package scraper

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"bitbucket.org/hilmarp/price-scraper/metrics"
	"github.com/robfig/cron/v3"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// StartWatcher will send price alerts
func (s *Scraper) StartWatcher() error {
	c := cron.New()
	c.AddFunc("20 */2 * * *", func() { // At minute 20 past every 2nd hour
		metrics.WatchersRunning.Inc()
		defer metrics.WatchersRunning.Dec()

		limit := 100
		offset := 0

		// From is 3 days
		now := time.Now()
		from := now.Add(time.Duration(-72) * time.Hour)

		absPath := os.Getenv("PRICE_ABS_PATH")

		var wg sync.WaitGroup

		for {
			watchProducts, err := s.DB.GetWatchProducts(limit, offset)
			if err != nil {
				log.Print(err)
				break
			}

			if len(*watchProducts) == 0 {
				break
			}

			for _, watchProduct := range *watchProducts {
				prices, err := s.DB.GetProductPrices(watchProduct.ProductID, from, "id desc")
				if err != nil {
					log.Print(err)
					continue
				}

				if len(*prices) == 0 {
					continue
				}

				currentPrice := (*prices)[0]

				for _, price := range *prices {
					// If the price is older than the watcher, we break
					if price.CreatedAt.Before(watchProduct.CreatedAt) {
						break
					}

					// Same price, keep going
					if price.Price == currentPrice.Price {
						continue
					}

					// Lower price found which means the product now has a higher price,
					// will be handled later, for now keep going
					if price.Price < currentPrice.Price {
						continue
					}

					// Higher price found which means the product has dropped in price, so notify the watcher
					if price.Price > currentPrice.Price {
						// If en email has already been sent about this price entry, break here
						if watchProduct.Sent != nil && *watchProduct.PriceIDSent == price.ID {
							break
						}

						wg.Add(1)
						go func(watchProduct WatchProduct, currentPrice, price Price) {
							defer wg.Done()

							type email struct {
								UnsubscribeHash string
								PriceOld        string
								PriceNew        string
								PriceDiff       string
								ProductURL      string
								ProductTitle    string
								Date            string
							}

							product, err := s.DB.GetProductByID(watchProduct.ProductID)
							if err != nil {
								log.Println(err)
								return
							}

							emailTxt := email{
								UnsubscribeHash: watchProduct.UnsubscribeHash,
								PriceOld:        strconv.Itoa(int(price.Price)),
								PriceNew:        strconv.Itoa(int(currentPrice.Price)),
								PriceDiff:       strconv.Itoa(int(price.Price - currentPrice.Price)),
								ProductURL:      fmt.Sprintf("https://verdfra.is/product/%v", product.Slug),
								ProductTitle:    product.Title,
								Date:            currentPrice.Date.Format("02/01/2006 15:04"),
							}

							tmpl, err := template.ParseFiles(fmt.Sprintf("%s/templates/watch-product.html", absPath))
							if err != nil {
								log.Println(err)
								return
							}

							var tmplBuffer bytes.Buffer
							err = tmpl.Execute(&tmplBuffer, emailTxt)
							if err != nil {
								log.Println(err)
								return
							}

							from := mail.NewEmail("Verð frá", os.Getenv("PRICE_EMAIL_FROM"))
							subject := fmt.Sprintf("Verðlækkun - %v", product.Title)
							to := mail.NewEmail(watchProduct.Email, watchProduct.Email)
							tmplStr := tmplBuffer.String()
							message := mail.NewSingleEmail(from, subject, to, tmplStr, tmplStr)
							client := sendgrid.NewSendClient(os.Getenv("PRICE_EMAIL_API_KEY"))
							_, err = client.Send(message)
							if err != nil {
								log.Println(err)
								return
							}

							metrics.WatcherEmailsSent.Inc()

							watchProduct.Sent = &now
							watchProduct.PriceIDSent = &price.ID
							err = s.DB.UpdateWatchProduct(&watchProduct)
							if err != nil {
								log.Println(err)
								return
							}
						}(watchProduct, currentPrice, price)

						// No need to check any more
						break
					}
				}
			}

			offset = offset + limit
		}

		wg.Wait()
	})
	c.Start()

	return nil
}
