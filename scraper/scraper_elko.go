package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) elkoCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "elko.is" {
		return
	}

	if e.DOM.Find(".product-page").Length() == 0 {
		return
	}

	title := e.ChildText("#product_title")
	code := e.ChildText(".product-detail-content .product-code")
	priceText := e.ChildText(".product-price-content .product-price")
	description := strings.TrimSpace(e.ChildText("#description"))

	// All images
	imgSrcs := e.ChildAttrs("#product_img_slider_content img", "data-src")
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
	e.ForEach(".feature-info table tr", func(_ int, el *colly.HTMLElement) {
		tds := el.ChildTexts("td")
		if len(tds) > 1 {
			specs = append(specs, Spec{Key: tds[0], Value: tds[1]})
		}
	})

	// Stock, Elko has more than 1 layout for stocks, check both
	trLayout := e.DOM.Find(".product-detail-content .stock-section table tr").Length() > 0
	pLayout := e.DOM.Find(".product-detail-content .stock-section p img").Length() > 0

	stocks := make([]Stock, 0)

	if trLayout {
		e.ForEach(".product-detail-content .stock-section table tr", func(_ int, el *colly.HTMLElement) {
			tds := el.ChildTexts("td")
			inStock := false
			if len(tds) > 2 {
				inStock = tds[2] == "Til á lager" || tds[2] == "Fá eintök eftir"
			}
			stocks = append(stocks, Stock{Location: tds[0], InStock: inStock})
		})
	}

	if pLayout {
		e.ForEach(".product-detail-content .stock-section p img", func(_ int, el *colly.HTMLElement) {
			imgSrc := el.Attr("src")
			inStock := strings.HasSuffix(imgSrc, "productview-low-stock.svg") || strings.HasSuffix(imgSrc, "productview-in-stock.svg")
			loc := ""

			el.DOM.Parent().Each(func(i int, s *goquery.Selection) {
				if i == 0 {
					loc = strings.Split(s.Text(), ":")[0]
				}
			})

			if loc != "" {
				stocks = append(stocks, Stock{Location: loc, InStock: inStock})
			}
		})
	}

	// Categories
	categories := getCategoriesFromBreadcrumbs(e.DOM.Find(".breadcrumb-content li a"), false, false)

	// Elko slug is special, the product might be on the root path
	// or /subcategory/something, so code and title could be duplicates in db
	slug := strings.ReplaceAll(productURL, "https://elko.is/", "")
	slugParts := strings.Split(slug, "/")
	slug = formatters.GetSlug("el", strings.Join(slugParts, "-"))

	product := Product{
		Source:      "elko.is",
		ProductCode: code,
		Slug:        slug,
		URL:         productURL,
		Title:       title,
		Description: description,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".product-price-content .product-discount").Length() > 0,
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
