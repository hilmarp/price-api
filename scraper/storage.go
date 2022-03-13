package scraper

import (
	"fmt"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"bitbucket.org/hilmarp/price-scraper/metrics"
	"github.com/gocolly/colly/v2/queue"
	"github.com/gocolly/colly/v2/storage"
)

// Storage describes what a scraper storage should do
type Storage interface {
	storage.Storage
	queue.Storage
	Clear() error
}

// StoreProduct saves product to database and elasticsearch
func (s *Scraper) StoreProduct(product *Product) error {
	// MySQL
	product.URL = formatters.GetCleanURL(formatters.GetURLWithoutWWW(product.URL), []string{"ProductID"})

	storedProduct, err := s.DB.UpdateOrCreateProduct(product)
	if err != nil {
		return fmt.Errorf("error storing product %s in database: %w", product.URL, err)
	}

	metrics.ProductStoredCount.Inc()

	// Elasticsearch
	categories := make([]string, len(product.Categories))
	for i, c := range product.Categories {
		categories[i] = c.Name
	}

	urls := []string{
		formatters.GetCleanURL(formatters.GetURLWithoutWWW(product.URL), []string{"ProductID"}),
		formatters.GetCleanURL(formatters.GetURLWithWWW(product.URL), []string{"ProductID"}),
	}

	searchProduct := &SearchProduct{
		ID:          storedProduct.ID,
		ScrapedAt:   time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Source:      product.Source,
		ProductCode: product.ProductCode,
		Slug:        product.Slug,
		URL:         urls,
		Title:       product.Title,
		Categories:  categories,
		Description: product.Description,
		MainImgURL:  product.MainImgURL,
		Price:       product.Price,
		OnSale:      product.OnSale,
	}
	err = s.ES.UpdateOrIndexSearchProduct(searchProduct)
	if err != nil {
		return fmt.Errorf("error storing search product in Elasticsearch: %w", err)
	}

	metrics.ProductStoredESCount.Inc()

	return nil
}
