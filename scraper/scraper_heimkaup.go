package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) heimkaupCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "heimkaup.is" {
		return
	}

	if strings.Contains(productURL, "heimkaup.is/share/") {
		return
	}

	title := e.ChildText(".Header-title h1")
	priceText := e.DOM.Find(".SideDetails-basket .Price-price .Price").Contents().Not("s").Text()

	if priceText == "" {
		priceText = e.ChildAttr(".SideDetails-serialPayments #netgiro-serial", "data-amount")
	}

	// Description
	description := strings.TrimSpace(e.ChildText(".ProductDetails-details.ProductDetails-section"))
	if description == "" {
		description = strings.TrimSpace(e.ChildText(".ProductDetails-description.ProductDetails-section"))
	}

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// Code
	code := e.ChildText(".SideDetails-brand .Details-partNumber")
	code = strings.ReplaceAll(code, "Vörunúmer: ", "")

	// Specs
	specs := make([]Spec, 0)
	e.ForEach(".ProductDetails-parameters.ProductDetails-section .list li", func(_ int, el *colly.HTMLElement) {
		texts := strings.Split(el.Text, ":")
		if len(texts) > 1 {
			specs = append(specs, Spec{Key: strings.TrimSpace(texts[0]), Value: strings.TrimSpace(texts[1])})
		}
	})

	// Stocks
	stocks := []Stock{{
		Location: "Vefverslun",
		InStock:  true,
	}}

	// All images
	imgSrcs := e.ChildAttrs(".ProductDetails-gallery.ProductDetails-section .swiper-wrapper.Gallery-images img", "src")
	allImgURLs := make([]Image, len(imgSrcs))
	for index, item := range imgSrcs {
		allImgURLs[index] = Image{URL: item, OriginalURL: item}
	}
	mainImgURL := ""
	if len(allImgURLs) > 0 {
		mainImgURL = allImgURLs[0].URL
	}

	// Categories
	categories := getCategoriesFromBreadcrumbs(e.DOM.ParentsUntil("~").Find("#snippet--headerBreadcrumbs ol li a"), true, true)

	product := &Product{
		Source:      "heimkaup.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("heimk", code, title),
		URL:         formatters.GetURLWithoutQuery(productURL),
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".SideDetails-basket .Price-discount").Length() > 0,
		Specs:       specs,
		Stocks:      stocks,
		AllImgURLs:  allImgURLs,
		Prices:      []Price{price},
		Categories:  categories,
	}

	err := s.StoreProduct(product)
	if err != nil {
		log.Println(err.Error())
	}
}
