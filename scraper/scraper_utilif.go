package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) utilifCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "utilif.is" {
		return
	}

	title := e.ChildText("h1.page-title")
	code := e.ChildText(".product.attribute.sku .value")
	priceText := e.ChildText(".product-info-price .normal-price .price")
	description := strings.TrimSpace(e.ChildText(".product.attribute.description .value"))

	if priceText == "" {
		priceText = e.ChildText("span.price")
	}

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// All images
	imgSrcs := e.ChildAttrs(".MagicToolboxContainer a.mt-thumb-switcher", "href")
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
	e.ForEach(".product-attribute-specs-table tr", func(_ int, el *colly.HTMLElement) {
		key := el.ChildText("th")
		val := el.ChildText("td")
		specs = append(specs, Spec{Key: key, Value: val})
	})

	// Stock
	inStock := e.DOM.Find(".product-info-stock-sku .stock.available").Length() > 0
	stocks := []Stock{{Location: "Vefverslun", InStock: inStock}}

	// Categories, utilif.is is weird, need to extract them from the URL
	categoriesArr := strings.Split(strings.ReplaceAll(productURL, "https://www.utilif.is/", ""), "/")
	if len(categoriesArr) > 0 {
		categoriesArr = categoriesArr[:len(categoriesArr)-1]
	}
	// Clean up, for example kk-utivistarfatnadur will become Kk utivistarfatnadur
	for i := 0; i < len(categoriesArr); i++ {
		cleaned := strings.Title(strings.ReplaceAll(categoriesArr[i], "-", " "))
		categoriesArr[i] = cleaned
	}
	categories := getCategoriesFromArray(categoriesArr)

	product := &Product{
		Source:      "utilif.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("ul", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".product-info-price .old-price").Length() > 0,
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
