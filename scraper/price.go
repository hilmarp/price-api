package scraper

import (
	"log"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"bitbucket.org/hilmarp/price-scraper/metrics"
	"github.com/robfig/cron/v3"
)

// StartPriceChangeWatcher checks if product price has changed
func (s *Scraper) StartPriceChangeWatcher() error {
	c := cron.New()
	c.AddFunc("50 */11 * * *", func() { // At minute 50 past every 11th hour.
		metrics.PriceChangeWatchersRunning.Inc()
		defer metrics.PriceChangeWatchersRunning.Dec()

		limit := 100
		offset := 0

		// From is two weeks
		now := time.Now()
		from := now.Add(time.Duration(-336) * time.Hour)

		for {
			products, err := s.DB.GetProducts(limit, offset, 0, 0, "id desc", "", []string{}, []string{})
			if err != nil {
				log.Print(err)
				break
			}

			if len(*products) == 0 {
				break
			}

			for _, product := range *products {
				prices, err := s.DB.GetProductPrices(product.ID, from, "id desc")
				if err != nil {
					log.Print(err)
					continue
				}

				if len(*prices) == 0 {
					continue
				}

				currentPrice := (*prices)[0]
				lenPrices := len(*prices)
				for i, price := range *prices {
					if i == 0 {
						continue
					}

					priceDiff := formatters.GetNumbersDiff(int(currentPrice.Price), int(price.Price))
					if priceDiff > 0 {
						err := s.DB.UpdateOrCreateProductPriceChange(&ProductPriceChange{
							ProductID:     product.ID,
							PriceDiff:     priceDiff,
							PriceLower:    currentPrice.Price < price.Price,
							PrevPriceDate: price.Date,
						})
						if err != nil {
							log.Print(err)
						}
						break
					}

					// If we get all the way to the end, set it to that price, which should be 0
					if i == lenPrices-1 {
						err := s.DB.UpdateOrCreateProductPriceChange(&ProductPriceChange{
							ProductID:     product.ID,
							PriceDiff:     priceDiff,
							PriceLower:    false,
							PrevPriceDate: price.Date,
						})
						if err != nil {
							log.Print(err)
						}
					}
				}
			}

			offset = offset + limit
		}
	})
	c.Start()

	return nil
}
