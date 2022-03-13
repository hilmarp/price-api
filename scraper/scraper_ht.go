package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) htCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "ht.is" {
		return
	}

	if e.DOM.Find(".product-info").Length() == 0 {
		return
	}

	title := e.ChildText("h1.product-title")
	code := e.ChildText(".product-nr")
	priceText := e.ChildText(".product-price")
	description := strings.TrimSpace(e.ChildText(".product-preDesc"))

	// All images
	imgSrcs := e.ChildAttrs(".owl-product-image .image-item img", "src")
	allImgURLs := make([]Image, len(imgSrcs))
	for index, item := range imgSrcs {
		allImgURLs[index] = Image{URL: item, OriginalURL: item}
	}
	mainImgURL := ""
	if len(allImgURLs) > 0 {
		mainImgURL = allImgURLs[0].URL
	}

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// Specs
	specs := make([]Spec, 0)
	e.ForEach(".specText table tr", func(_ int, el *colly.HTMLElement) {
		tds := el.ChildTexts("td")
		if len(tds) > 1 {
			specs = append(specs, Spec{Key: tds[0], Value: tds[1]})
		}
	})

	// Stock
	stocks := make([]Stock, 0)
	e.ForEach(".stores-status li", func(_ int, el *colly.HTMLElement) {
		loc := el.ChildText(".storeStatus-title")
		inStock := el.DOM.Find(".glyphicon.glyphicon-ok").Length() > 0
		stocks = append(stocks, Stock{Location: loc, InStock: inStock})
	})

	// Categories
	categories := getCategoriesFromBreadcrumbs(e.DOM.ParentsUntil("~").Find(".breadcrumbs a"), true, false)

	product := Product{
		Source:      "ht.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("ht", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		Price:       price.Price,
		OnSale:      e.DOM.Find("#product .discount-percent").Length() > 0,
		Specs:       specs,
		Stocks:      stocks,
		AllImgURLs:  allImgURLs,
		MainImgURL:  mainImgURL,
		Prices:      []Price{price},
		Categories:  categories,
	}

	err := s.StoreProduct(&product)
	if err != nil {
		log.Println(err.Error())
	}
}
