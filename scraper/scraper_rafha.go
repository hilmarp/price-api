package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) rafhaCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "rafha.is" {
		return
	}

	title := e.ChildText("h1.product_title")
	code := e.ChildText(".loop-product-categories a")
	priceText := e.ChildText(".electro-price ins .amount")
	description := strings.TrimSpace(e.ChildText(".electro-description"))

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// All images
	imgSrcs := e.ChildAttrs(".thumbnails-single.owl-carousel a", "href")
	allImgURLs := make([]Image, len(imgSrcs))
	for index, item := range imgSrcs {
		allImgURLs[index] = Image{URL: item, OriginalURL: item}
	}
	mainImgURL := ""
	if len(allImgURLs) > 0 {
		mainImgURL = allImgURLs[0].URL
	}

	// Stock
	stock := e.ChildText(".availability span")
	stocks := []Stock{
		{Location: "Suðurlandsbraut", InStock: stock == "Til á lager"},
	}

	// Categories
	categories := getCategoriesFromBreadcrumbs(e.DOM.ParentsUntil("~").Find(".woocommerce-breadcrumb a"), false, true)

	product := Product{
		Source:      "rafha.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("rh", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".electro-price del").Length() > 0,
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
