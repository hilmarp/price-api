package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) rumfatalagerinnCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "rumfatalagerinn.is" {
		return
	}

	title := e.ChildText("h1.new-product-main-header")
	description := strings.TrimSpace(e.ChildText(".new-product-description-text-container"))
	priceText := e.ChildText(".new-product-price__price")

	// Code
	code := e.ChildText(".new-product-main-title p")
	code = strings.ReplaceAll(code, "Vörunúmer: ", "")

	// All images
	imgSrcs := e.ChildAttrs("#img-slide-row a.img-slide", "href")
	allImgURLs := make([]Image, len(imgSrcs))
	for i, src := range imgSrcs {
		absSrc := e.Request.AbsoluteURL(src)
		allImgURLs[i] = Image{URL: absSrc, OriginalURL: absSrc}
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
	e.ForEach(".new-properties-container .new-property-item", func(_ int, el *colly.HTMLElement) {
		tds := el.ChildTexts("p")
		if len(tds) > 1 {
			specs = append(specs, Spec{Key: tds[0], Value: tds[1]})
		}
	})

	// Stock
	stocks := make([]Stock, 0)
	e.ForEach(".availability-list li", func(_ int, el *colly.HTMLElement) {
		loc := el.DOM.Contents().Not("span").Text()
		inStock := el.DOM.HasClass("available")
		stocks = append(stocks, Stock{Location: loc, InStock: inStock})
	})

	// Categories
	categories := getCategoriesFromBreadcrumbs(e.DOM.ParentsUntil("~").Find(".breadcrumbs-container a"), false, true)

	product := Product{
		Source:      "rumfatalagerinn.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("rumf", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".new-product-price__offer-price strike").Length() > 0,
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
