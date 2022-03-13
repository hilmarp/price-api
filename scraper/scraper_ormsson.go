package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) ormssonCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "ormsson.is" {
		return
	}

	productDOM := e.DOM.ParentsUntil("~").Find("#rightbar")

	title := productDOM.Find("h1.h1-text").Text()
	priceText := productDOM.Find(".thisprice").Not(".oldPrice").Text()
	description := strings.TrimSpace(productDOM.Find(".precontent").Children().Text())

	// Product code, looks like "vrn. SAQE55Q95TATXXC"
	code := productDOM.Find(".productNr").Text()
	code = strings.ReplaceAll(code, "vrn. ", "")

	// Price at date
	price := Price{
		Price: formatters.StringToPrice(priceText),
		Date:  time.Now(),
	}

	// Images
	imgs := productDOM.Find(".col-lg-5.col-md-6.col-sm-12 a[data-lightbox] img")
	allImgURLs := make([]Image, 0)
	imgs.Each(func(_ int, s *goquery.Selection) {
		src := s.AttrOr("src", "")
		if len(src) < 256 {
			allImgURLs = append(allImgURLs, Image{URL: src, OriginalURL: src})
		}
	})
	mainImgURL := ""
	if len(allImgURLs) > 0 {
		mainImgURL = allImgURLs[0].URL
	}

	// Specs
	specs := make([]Spec, 0)
	productDOM.Find(".pcontent table tr").Each(func(_ int, s *goquery.Selection) {
		key := strings.TrimSpace(s.Find("td").First().Text())
		val := strings.TrimSpace(s.Find("td").Last().Text())
		if len(key) < 256 && len(val) < 256 {
			specs = append(specs, Spec{Key: key, Value: val})
		}
	})

	// Stocks
	stocks := make([]Stock, 0)
	productDOM.Find(".warehouses ul li").Each(func(_ int, s *goquery.Selection) {
		stocks = append(stocks, Stock{Location: s.Text(), InStock: s.HasClass("true")})
	})

	// Categories
	categories := getCategoriesFromBreadcrumbs(
		productDOM.Find("ul.breadcrumb").First().Find("a"),
		false,
		true,
	)

	product := &Product{
		Source:      "ormsson.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("orm", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      productDOM.Find(".oldPrice").Length() > 0,
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
