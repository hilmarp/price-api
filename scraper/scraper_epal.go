package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) epalCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "epal.is" {
		return
	}

	title := e.ChildText("h1.product-title")
	code := e.ChildText(".product_meta .sku_wrapper .sku")
	priceText := e.ChildText(".price-wrapper .amount")
	description := strings.TrimSpace(e.ChildText("#tab-description p"))

	// Images
	imgSrcs := e.ChildAttrs(".product-thumbnails.thumbnails img.attachment-woocommerce_thumbnail", "src")
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
	e.ForEach("#tab-description table tr", func(_ int, el *colly.HTMLElement) {
		key := el.ChildText("th")
		val := el.ChildText("td")
		specs = append(specs, Spec{Key: key, Value: val})
	})

	// Stock
	stocks := make([]Stock, 0)
	e.ForEach(".warehouse-info .warehouse-items li", func(_ int, el *colly.HTMLElement) {
		loc := el.Text
		inStock := el.DOM.Find("i.fa-check").Length() > 0
		stocks = append(stocks, Stock{Location: loc, InStock: inStock})
	})

	// Categories
	categoriesArr := make([]string, 0)
	e.ForEach(".woocommerce-breadcrumb.breadcrumbs a", func(i int, el *colly.HTMLElement) {
		if i > 1 {
			categoriesArr = append(categoriesArr, el.Text)
		}
	})
	categories := getCategoriesFromArray(categoriesArr)

	product := Product{
		Source:      "epal.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("ep", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".price-wrapper .price-on-sale").Length() > 0,
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
