package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) penninnCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "penninn.is" {
		return
	}

	if e.DOM.Find(".gtm-details").Length() == 0 {
		return
	}

	title := e.ChildText(".gtm-details h1.hdln--larger")
	description := strings.TrimSpace(e.ChildText(".field-type-text-with-summary p"))
	priceText := e.ChildText(".commerce-price-savings-formatter-price .price-amount")

	// Code
	code := e.ChildText(".prod-num")
	code = strings.ReplaceAll(code, "Vörunúmer: ", "")

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// Stock
	stocks := make([]Stock, 0)
	e.ForEach(".locations__container ul li", func(_ int, el *colly.HTMLElement) {
		stocks = append(stocks, Stock{Location: el.Text, InStock: true})
	})

	// Images
	imgSrcs := e.ChildAttrs(".my-gallery figure a", "href")
	allImgURLs := make([]Image, len(imgSrcs))
	for index, item := range imgSrcs {
		allImgURLs[index] = Image{URL: item, OriginalURL: item}
	}
	mainImgURL := ""
	if len(allImgURLs) > 0 {
		mainImgURL = allImgURLs[0].URL
	}

	// Categories, penninn.is is weird, need to extract them from the URL
	categoriesArr := strings.Split(strings.ReplaceAll(productURL, "https://www.penninn.is/", ""), "/")
	if len(categoriesArr) > 1 {
		categoriesArr = categoriesArr[:len(categoriesArr)-1]
		categoriesArr = categoriesArr[1:]
	}
	// Clean up, for example husgogn will become Husgogn
	for i := 0; i < len(categoriesArr); i++ {
		cleaned := strings.Title(strings.ReplaceAll(categoriesArr[i], "-", " "))
		categoriesArr[i] = cleaned
	}
	categories := getCategoriesFromArray(categoriesArr)

	product := &Product{
		Source:      "penninn.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("penn", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".commerce-price-savings-formatter-savings").Length() > 0,
		Specs:       []Spec{},
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
