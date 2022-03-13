package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) nexusCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "nexus.is" {
		return
	}

	title := e.ChildText("h2.single-post-title")
	code := e.ChildAttr(".tinv-wraper.woocommerce.tinv-wishlist.tinvwl-before-add-to-cart", "data-product_id")
	description := strings.TrimSpace(e.ChildText("div#tab-description"))
	priceText := e.ChildText(".summary.entry-summary .woocommerce-Price-amount.amount")

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// All images
	imgSrcs := e.ChildAttrs(".woocommerce-product-gallery .woocommerce-product-gallery__image img", "src")
	allImgURLs := make([]Image, len(imgSrcs))
	for index, item := range imgSrcs {
		allImgURLs[index] = Image{URL: item, OriginalURL: item}
	}
	mainImgURL := ""
	if len(allImgURLs) > 0 {
		mainImgURL = allImgURLs[0].URL
	}

	// Stocks
	stocks := []Stock{{
		Location: "Vefverslun",
		InStock:  e.DOM.Find(".stock.in-stock").Length() > 0,
	}}

	// Specs
	specs := make([]Spec, 0)
	e.ForEach("div#tab-additional_information table tr", func(_ int, el *colly.HTMLElement) {
		key := el.ChildText("th")
		val := el.ChildText("td")
		specs = append(specs, Spec{Key: key, Value: val})
	})

	// Categories
	categoryURL := e.ChildAttr(".posted_in a", "href")
	categoryURL = strings.ReplaceAll(categoryURL, "https://nexus.is/voruflokkur/", "")
	categoryURL = strings.TrimSuffix(categoryURL, "/")
	categoryURLs := strings.Split(categoryURL, "/")
	categories := getCategoriesFromArray(categoryURLs)

	product := Product{
		Source:      "nexus.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("nex", code, title),
		URL:         formatters.GetURLWithoutQuery(productURL),
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      false,
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
