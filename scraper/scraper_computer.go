package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) computerCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "computer.is" {
		return
	}

	title := e.ChildText("h2.header-title")
	priceText := e.ChildText(".pantavoru .displayPrice")
	description := strings.TrimSpace(e.ChildText(".preContent"))

	// Code
	code := ""
	e.ForEach(".pantavoru .extraInfo", func(_ int, el *colly.HTMLElement) {
		if strings.Contains(el.Text, "Framl.númer: ") {
			code = strings.Split(el.Text, " ")[1]
		}
	})
	if code == "" {
		log.Printf("Could not find product code for %s\n", productURL)
		return
	}

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// All images
	imgSrcs := e.ChildAttrs(".productImg .product-image-main a", "href")
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
	e.ForEach(".product-desc.visible-md.visible-lg ul li", func(_ int, el *colly.HTMLElement) {
		split := strings.Split(el.Text, ": ")
		if len(split) > 1 && len(split[0]) < 256 && len(split[1]) < 256 {
			specs = append(specs, Spec{Key: split[0], Value: split[1]})
		}
	})

	// Stock
	stock := e.ChildText(".status .status-text")
	stocks := []Stock{
		{Location: "Skipholt", InStock: stock == "Til á lager"},
	}

	// Categories
	var categories []Category
	breadcrumbs := e.DOM.ParentsUntil("~").Find(".crumbs .breadcrumb")
	if breadcrumbs.Length() > 1 {
		categories = getCategoriesFromBreadcrumbs(breadcrumbs.Last().Find("a"), false, true)
	} else {
		categories = getCategoriesFromBreadcrumbs(breadcrumbs.First().Find("a"), false, true)
	}

	product := Product{
		Source:      "computer.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("comp", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".discountRibbon").Length() > 0,
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
