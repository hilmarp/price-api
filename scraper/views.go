package scraper

import (
	"log"
	"time"

	"bitbucket.org/hilmarp/price-scraper/metrics"
	"github.com/robfig/cron/v3"
)

// StartViewCounter handles view counter stuff
func (s *Scraper) StartViewCounter() error {
	c := cron.New()
	c.AddFunc("30 */3 * * *", func() { // At minute 30 past every 3rd hour.
		metrics.ViewCountersRunning.Inc()
		defer metrics.ViewCountersRunning.Dec()

		limit := 100
		offset := 0

		// Before will be deleted
		now := time.Now()
		before := now.Add(time.Duration(-336) * time.Hour) // 2 weeks

		for {
			productViewCounts, err := s.DB.GetProductViewCounts(limit, offset)
			if err != nil {
				log.Print(err)
				break
			}

			if len(*productViewCounts) == 0 {
				break
			}

			for _, productViewCount := range *productViewCounts {
				if productViewCount.CreatedAt.Before(before) {
					err = s.DB.DeleteProductViewCountByID(productViewCount.ID)
					if err != nil {
						log.Print(err)
						continue
					}
				}
			}

			offset = offset + limit
		}
	})
	c.Start()

	return nil
}
