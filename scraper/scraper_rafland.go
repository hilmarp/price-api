package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) raflandCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "rafland.is" {
		return
	}

	if e.DOM.Find(".product-head").Length() == 0 {
		return
	}

	title := e.ChildText(".product-title h1")
	code := e.ChildText(".product-nr")
	priceText := e.ChildText(".addtocartform .btn-cart")

	// Description and specs, it has the same layout
	description := ""
	specs := make([]Spec, 0)

	e.ForEach(".product-body .product-body-item", func(index int, el *colly.HTMLElement) {
		// Description
		if index == 0 {
			description = strings.TrimSpace(el.ChildText(".product-body-content .col-md-8"))
		}

		// Specs
		if index == 1 {
			el.ForEach(".specText table tr", func(_ int, ele *colly.HTMLElement) {
				tds := ele.ChildTexts("td")
				spec := Spec{}

				if len(tds) > 1 {
					spec.Key = tds[0]
					spec.Value = tds[1]
				}

				if len(tds) > 2 && tds[1] == "" {
					spec.Value = tds[2]
				}

				specs = append(specs, spec)
			})
		}
	})

	// Code
	if strings.Contains(code, ":") {
		code = strings.Split(code, ":")[1]
		code = strings.TrimSpace(code)
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

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// Stock
	stock := e.ChildText(".product-status")
	stocks := []Stock{
		{Location: "Síðumúla", InStock: stock == "Á Lager"},
	}

	// Categories
	categories := getCategoriesFromBreadcrumbs(e.DOM.ParentsUntil("~").Find(".breadcrumbs a"), true, false)

	product := Product{
		Source:      "rafland.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("rl", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		Price:       price.Price,
		OnSale:      e.DOM.Find("#product .old-price").Length() > 0,
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
