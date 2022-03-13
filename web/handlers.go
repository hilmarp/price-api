package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"bitbucket.org/hilmarp/price-scraper/metrics"
	"bitbucket.org/hilmarp/price-scraper/scraper"
	"github.com/go-chi/chi"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func (s *APIServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi!"))
}

func (s *APIServer) productHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	product, err := s.DB.GetProductBySlug(slug)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Increment view counter
	err = s.DB.IncrementProductViewCount(product.ID)
	if err != nil {
		log.Printf("Error increment product view counter: %s", err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func (s *APIServer) productPricesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	orderByDirQ := r.URL.Query().Get("order_by_dir")
	orderByDir := "asc"
	if orderByDirQ != "" && orderByDirQ == "desc" {
		orderByDir = orderByDirQ
	}
	orderBy := fmt.Sprintf("id %s", orderByDir)

	// From date
	var from time.Time
	fromQ := r.URL.Query().Get("from")
	now := time.Now()
	if fromQ == "" {
		// Default to 30 days
		from = now.Add(time.Duration(-720) * time.Hour)
	} else {
		parsed, err := time.Parse("2006-01-02", fromQ)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// Set max date, 1 year
		max := now.Add(time.Duration(-8760) * time.Hour)

		if parsed.Before(max) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Date set to far back in time"))
			return
		}

		from = parsed
	}

	prices, err := s.DB.GetProductPrices(uint(id), from, orderBy)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prices)
}

func (s *APIServer) productSpecsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	specs, err := s.DB.GetProductSpecs(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(specs)
}

func (s *APIServer) productStocksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	stocks, err := s.DB.GetProductStocks(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stocks)
}

func (s *APIServer) productImagesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	images, err := s.DB.GetProductImages(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

func (s *APIServer) productCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	categories, err := s.DB.GetProductCategories(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (s *APIServer) productsHandler(w http.ResponseWriter, r *http.Request) {
	limitQ := r.URL.Query().Get("limit")
	offsetQ := r.URL.Query().Get("offset")
	orderBy := r.URL.Query().Get("order_by")
	orderByDir := r.URL.Query().Get("order_by_dir")
	sourcesQ := r.URL.Query().Get("sources")
	categorySlugsQ := r.URL.Query().Get("categories")
	priceFromQ := r.URL.Query().Get("price_from")
	priceToQ := r.URL.Query().Get("price_to")
	onSaleQ := r.URL.Query().Get("on_sale")

	// Set defaults, then override if needed
	maxLimit := 100
	limit := maxLimit
	offset := 0
	priceFrom := 0
	priceTo := 0
	onSale := ""

	if priceFromQ != "" {
		num, err := strconv.Atoi(priceFromQ)
		if err == nil {
			priceFrom = num
		}
	}

	if priceToQ != "" {
		num, err := strconv.Atoi(priceToQ)
		if err == nil {
			priceTo = num
		}
	}

	if limitQ != "" {
		num, err := strconv.Atoi(limitQ)
		if err == nil {
			limit = num
		}
	}

	if offsetQ != "" {
		num, err := strconv.Atoi(offsetQ)
		if err == nil {
			offset = num
		}
	}

	if onSaleQ != "" {
		if onSaleQ == "true" || onSaleQ == "false" {
			onSale = onSaleQ
		}
	}

	// Final order by combining orderBy and orderByDir
	// ex. "price desc"
	order := "id desc"
	if orderBy != "" && orderBy == "price" {
		ordDir := "asc"
		if orderByDir != "" && orderByDir == "desc" {
			ordDir = orderByDir
		}
		order = fmt.Sprintf("%s %s", orderBy, ordDir)
	}

	// Don't want to crash the server by returning too many products
	if limit > maxLimit {
		limit = maxLimit
	}

	// Sources, separated by comma
	var sources []string
	if sourcesQ != "" {
		sources = strings.Split(sourcesQ, ",")
	}

	// Category slugs, separated by comma
	var categorySlugs []string
	if categorySlugsQ != "" {
		categorySlugs = strings.Split(categorySlugsQ, ",")
	}

	products, err := s.DB.GetProducts(limit, offset, priceFrom, priceTo, order, onSale, sources, categorySlugs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (s *APIServer) gotoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	product, err := s.DB.GetProductByID(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	err = s.DB.CreateProductClickCount(&scraper.ProductClickCount{ProductID: product.ID})
	if err != nil {
		log.Printf("Error creating product click count: %s", err.Error())
	}

	gotoURL := formatters.GetURLWithQueryParam(product.URL, "utm_source", "verdfra.is")

	metrics.ProductClickCount.Inc()

	http.Redirect(w, r, gotoURL, http.StatusSeeOther)
}

func (s *APIServer) searchHandler(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("value")
	limitQ := r.URL.Query().Get("limit")
	offsetQ := r.URL.Query().Get("offset")

	if value == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No value"))
		return
	}

	// Set defaults
	maxLimit := 100
	limit := maxLimit
	offset := 0

	if limitQ != "" {
		num, err := strconv.Atoi(limitQ)
		if err == nil {
			limit = num
		}
	}

	if offsetQ != "" {
		num, err := strconv.Atoi(offsetQ)
		if err == nil {
			offset = num
		}
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	esProducts, err := s.ES.SearchForProduct(value, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	metrics.TotalSearches.Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(esProducts)
}

func (s *APIServer) productsCountHandler(w http.ResponseWriter, r *http.Request) {
	sourcesQ := r.URL.Query().Get("sources")
	categorySlugsQ := r.URL.Query().Get("categories")
	priceFromQ := r.URL.Query().Get("price_from")
	priceToQ := r.URL.Query().Get("price_to")
	onSaleQ := r.URL.Query().Get("on_sale")

	// Set defaults
	priceFrom := 0
	priceTo := 0
	onSale := ""

	if priceFromQ != "" {
		num, err := strconv.Atoi(priceFromQ)
		if err == nil {
			priceFrom = num
		}
	}

	if priceToQ != "" {
		num, err := strconv.Atoi(priceToQ)
		if err == nil {
			priceTo = num
		}
	}

	if onSaleQ != "" {
		if onSaleQ == "true" || onSaleQ == "false" {
			onSale = onSaleQ
		}
	}

	// Sources, separated by comma
	var sources []string
	if sourcesQ != "" {
		sources = strings.Split(sourcesQ, ",")
	}

	// Category slugs, separated by comma
	var categorySlugs []string
	if categorySlugsQ != "" {
		categorySlugs = strings.Split(categorySlugsQ, ",")
	}

	count, err := s.DB.GetProductsCount(0, 0, priceFrom, priceTo, "", onSale, sources, categorySlugs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	type countJSON struct {
		Count int
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(countJSON{Count: count})
}

func (s *APIServer) productsPopularHandler(w http.ResponseWriter, r *http.Request) {
	limitQ := r.URL.Query().Get("limit")
	offsetQ := r.URL.Query().Get("offset")

	// Set defaults
	maxLimit := 100
	limit := maxLimit
	offset := 0

	if limitQ != "" {
		num, err := strconv.Atoi(limitQ)
		if err == nil {
			limit = num
		}
	}

	if offsetQ != "" {
		num, err := strconv.Atoi(offsetQ)
		if err == nil {
			offset = num
		}
	}

	// Don't go over max limit
	if limit > maxLimit {
		limit = maxLimit
	}

	products, err := s.DB.GetPopularProducts(limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (s *APIServer) productsPriceChangesHandler(w http.ResponseWriter, r *http.Request) {
	limitQ := r.URL.Query().Get("limit")
	offsetQ := r.URL.Query().Get("offset")
	lower := r.URL.Query().Get("lower")

	// Set defaults
	maxLimit := 100
	limit := maxLimit
	offset := 0

	if limitQ != "" {
		num, err := strconv.Atoi(limitQ)
		if err == nil {
			limit = num
		}
	}

	if offsetQ != "" {
		num, err := strconv.Atoi(offsetQ)
		if err == nil {
			offset = num
		}
	}

	// Don't go over max limit
	if limit > maxLimit {
		limit = maxLimit
	}

	products, err := s.DB.GetProductsPriceChanges(limit, offset, lower)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (s *APIServer) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	parent := r.URL.Query().Get("parent")

	categories, err := s.DB.GetUniqueCategories(parent)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (s *APIServer) categoryHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	category, err := s.DB.GetUniqueCategoryBySlug(slug)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (s *APIServer) watchProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	product, err := s.DB.GetProductByID(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	type input struct {
		Email string
	}

	var in input
	err = json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if in.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no email"))
		return
	}

	verifyHash := formatters.GetRandomStringWithTimestamp(15)
	unsubscribeHash := formatters.GetRandomStringWithTimestamp(15)

	err = s.DB.CreateWatchProduct(&scraper.WatchProduct{
		Email:           in.Email,
		ProductID:       product.ID,
		Sent:            nil,
		PriceIDSent:     nil,
		Verified:        false,
		VerifyHash:      verifyHash,
		UnsubscribeHash: unsubscribeHash,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	type email struct {
		VerifyHash   string
		ProductTitle string
	}

	absPath := os.Getenv("PRICE_ABS_PATH")
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/templates/watch-verify.html", absPath))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var tmplBuffer bytes.Buffer
	err = tmpl.Execute(&tmplBuffer, email{VerifyHash: verifyHash, ProductTitle: product.Title})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	from := mail.NewEmail("Verð frá", os.Getenv("PRICE_EMAIL_FROM"))
	subject := "Staðfesta netfang"
	to := mail.NewEmail(in.Email, in.Email)
	tmplStr := tmplBuffer.String()
	message := mail.NewSingleEmail(from, subject, to, tmplStr, tmplStr)
	client := sendgrid.NewSendClient(os.Getenv("PRICE_EMAIL_API_KEY"))
	_, err = client.Send(message)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	metrics.WatcherSignups.Inc()

	w.Write([]byte("Watch!"))
}

func (s *APIServer) watchVerifyHandler(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no hash"))
		return
	}

	watchProduct, err := s.DB.GetWatchProductByVerifyHash(hash)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	watchProduct.Verified = true
	err = s.DB.UpdateWatchProduct(watchProduct)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	metrics.WatcherVerifies.Inc()

	w.Write([]byte("Verified!"))
}

func (s *APIServer) watchUnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no hash"))
		return
	}

	err := s.DB.DeleteWatchProductByUnsubscribeHash(hash)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	metrics.WatcherUnsubscribes.Inc()

	w.Write([]byte("Unsubscribed!"))
}

func (s *APIServer) contactHandler(w http.ResponseWriter, r *http.Request) {
	type email struct {
		From    string
		Message string
	}

	var e email
	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	absPath := os.Getenv("PRICE_ABS_PATH")
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/templates/contact.html", absPath))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var tmplBuffer bytes.Buffer
	err = tmpl.Execute(&tmplBuffer, e)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	from := mail.NewEmail("Verð frá", os.Getenv("PRICE_EMAIL_FROM"))
	subject := "Hafa samband"
	to := mail.NewEmail("Hilmar", "hilmar@hilmarp.com")
	tmplStr := tmplBuffer.String()
	message := mail.NewSingleEmail(from, subject, to, tmplStr, tmplStr)
	client := sendgrid.NewSendClient(os.Getenv("PRICE_EMAIL_API_KEY"))
	_, err = client.Send(message)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("Sent!"))
}

// productImageHandler loads an image from file and serves it
func (s *APIServer) productImageHandler(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")
	id := chi.URLParam(r, "id")
	absPath := os.Getenv("PRICE_ABS_PATH")
	path := fmt.Sprintf("%s/static/img/products/%s/%s.jpg", absPath, source, id)

	file, err := ioutil.ReadFile(path)
	if err != nil {
		// return placeholder
		file, err = ioutil.ReadFile(fmt.Sprintf("%s/static/img/products/placeholder.jpg", absPath))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("no such file or directory"))
			return
		}
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(file)
}

func (s *APIServer) productPriceChangeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	priceChange, err := s.DB.GetProductPriceChangeByProductID(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(priceChange)
}

func (s *APIServer) productsLastUpdatedHandler(w http.ResponseWriter, r *http.Request) {
	product, err := s.DB.GetLastUpdatedProduct()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
