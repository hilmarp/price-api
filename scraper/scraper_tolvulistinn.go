package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) tolvulistinnCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "tl.is" {
		return
	}

	title := e.ChildText(".product-title h1")
	priceText := e.ChildText(".addtocartform .btn-cart")
	description := strings.TrimSpace(e.DOM.Find(".product-body-content").First().Text())

	// Code
	code := e.ChildText(".product-nr")
	code = strings.ReplaceAll(code, "Vörunúmer : ", "")

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// All images
	imgSrcs := e.ChildAttrs(".product-image-slider .image-item img", "src")
	allImgURLs := make([]Image, len(imgSrcs))
	for index, item := range imgSrcs {
		allImgURLs[index] = Image{URL: item, OriginalURL: item}
	}
	mainImgURL := ""
	if len(allImgURLs) > 0 {
		mainImgURL = allImgURLs[0].URL
	}

	// Specs
	specs := make([]Spec, 0)
	e.ForEach(".product-body-content.specText table tr", func(_ int, el *colly.HTMLElement) {
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
		Source:      "tl.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("tl", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".old-price").Length() > 0,
		Specs:       specs,
		Stocks:      stocks,
		AllImgURLs:  allImgURLs,
		Prices:      []Price{price},
		Categories:  categories,
	}

	err := s.StoreProduct(&product)
	if err != nil {
		log.Println(err.Error())
	}
}
