package scraper

import (
	"log"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/gocolly/colly/v2"
)

func (s *Scraper) bykoCallback(e *colly.HTMLElement) {
	productURL := e.Request.URL.String()
	if formatters.GetURLHost(productURL) != "byko.is" {
		return
	}

	titleDom := e.DOM.ParentsUntil("~").Find(".productDetails_MainInformation_ProductName")
	if titleDom.Length() == 0 {
		return
	}

	title := titleDom.Text()
	code := e.ChildText(".productDetails_MainInformation_ProductNumber")
	priceText := e.ChildText(".productDetails_MainInformation_Price .priceTag_Price")
	description := strings.TrimSpace(e.ChildText(".productDetails__descriptionContainer"))

	// Images
	imgSrcs := e.ChildAttrs(".productDetails_Carousel .productDetails_Carousel_Item img", "src")
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
	e.ForEach(".productDetails_informationContainer table tr", func(_ int, el *colly.HTMLElement) {
		tds := el.ChildTexts("td")
		if len(tds) > 1 {
			specs = append(specs, Spec{Key: tds[0], Value: tds[1]})
		}
	})

	// Stocks
	stocks := []Stock{{
		Location: "Vefverslun",
		InStock:  e.DOM.Find("#notInStoreReason").Length() == 0,
	}}

	// Categories
	breadcrumbs := e.DOM.ParentsUntil("~").Find(".detail_Breadcrumb_OuterContainer").First()
	categories := getCategoriesFromBreadcrumbs(breadcrumbs.Find("a"), false, true)

	product := Product{
		Source:      "byko.is",
		ProductCode: code,
		Slug:        formatters.GetSlug("byk", code, title),
		URL:         productURL,
		Title:       title,
		Description: description,
		MainImgURL:  mainImgURL,
		Price:       price.Price,
		OnSale:      e.DOM.Find(".crashOverOldPrice").Length() > 0,
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
