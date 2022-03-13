package scraper

import (
	"time"

	"gorm.io/gorm"
)

// Product is the base product
type Product struct {
	gorm.Model
	Source      string `gorm:"index:idx_products_source_product_code"`
	ProductCode string `gorm:"index:idx_products_source_product_code"`
	Slug        string `gorm:"unique;size:255"`
	URL         string `gorm:"unique"`
	Title       string
	Description string `gorm:"type:text"`
	MainImgURL  string
	Price       uint // Latest price
	OnSale      bool
	Specs       []Spec
	Stocks      []Stock
	AllImgURLs  []Image
	Prices      []Price
	Categories  []Category
}

// SearchProduct is searchable fields in Elasticsearch
type SearchProduct struct {
	ID          uint
	ScrapedAt   string
	Source      string
	ProductCode string
	Slug        string
	URL         []string
	Title       string
	Categories  []string
	Description string
	MainImgURL  string
	Price       uint // latest price
	OnSale      bool
}

// ProductPriceDiff is the base product with price diff
type ProductPriceDiff struct {
	Product
	PriceDiff  uint
	PriceLower bool
}

type Price struct {
	gorm.Model
	Price     uint
	Date      time.Time
	ProductID uint `gorm:"index"`
}

type Image struct {
	gorm.Model
	URL         string
	OriginalURL string
	ProductID   uint `gorm:"index"`
}

type Stock struct {
	gorm.Model
	Location  string
	InStock   bool
	ProductID uint `gorm:"index"`
}

type Spec struct {
	gorm.Model
	Key       string
	Value     string
	ProductID uint `gorm:"index"`
}

type Category struct {
	gorm.Model
	Name      string
	Slug      string
	Parent    string
	ProductID uint `gorm:"index"`
}

// UniqueCategory has the same info as Category, but no duplicates
type UniqueCategory struct {
	gorm.Model
	Name   string
	Slug   string
	Parent string
}

// WatchProduct is an email watching a product
type WatchProduct struct {
	gorm.Model
	Email           string
	ProductID       uint
	Sent            *time.Time
	PriceIDSent     *uint
	Verified        bool   `gorm:"index"`
	VerifyHash      string `gorm:"unique"`
	UnsubscribeHash string `gorm:"unique"`
}

// ProductViewCount is a product web page view counter
type ProductViewCount struct {
	gorm.Model
	ProductID uint `gorm:"index"`
	Views     int
}

// ProductClickCount is a product external url click counter
type ProductClickCount struct {
	gorm.Model
	ProductID uint `gorm:"index"`
}

// ProductPriceChange contains price changes for product
type ProductPriceChange struct {
	gorm.Model
	ProductID     uint `gorm:"index"`
	PriceDiff     int
	PriceLower    bool
	PrevPriceDate time.Time
}

// Bot describes a website scraper robot
type Bot struct {
	gorm.Model
	URL        string `gorm:"index"`
	StartedAt  time.Time
	FinishedAt *time.Time
}
