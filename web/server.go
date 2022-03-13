package web

import (
	"fmt"
	"net/http"

	"bitbucket.org/hilmarp/price-scraper/scraper"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// APIServer is the /api web server
type APIServer struct {
	DB    *scraper.SQL
	ES    *scraper.Elasticsearch
	Redis *Redis
	Port  string
}

// StartServer will start the web server at localhost:port
func (s *APIServer) StartServer() error {
	r := chi.NewRouter()

	// Middleware
	// r.Use(middleware.Logger)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://verdfra.is", "https://www.verdfra.is", "http://localhost:3004"},
		AllowedMethods: []string{"GET", "HEAD", "PUT", "PATCH", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		MaxAge:         600,
	}))

	r.Get("/", s.rootHandler)
	r.Handle("/prometheus/metrics", promhttp.Handler())

	// Products
	r.Get("/product/{slug}", s.productHandler)
	r.Get("/product/{id}/prices", s.productPricesHandler)
	r.Get("/product/{id}/specs", s.productSpecsHandler)
	r.Get("/product/{id}/stocks", s.productStocksHandler)
	r.Get("/product/{id}/images", s.productImagesHandler)
	r.Get("/product/{id}/categories", s.productCategoriesHandler)
	r.Get("/product/{id}/price-change", s.productPriceChangeHandler)
	r.Get("/products", s.productsHandler)
	r.Get("/products/count", s.productsCountHandler)
	r.Get("/products/popular", s.productsPopularHandler)
	r.Get("/products/price-changes", s.productsPriceChangesHandler)
	r.Get("/products/last-updated", s.productsLastUpdatedHandler)

	// Categories
	r.Get("/categories", s.categoriesHandler)
	r.Get("/category/{slug}", s.categoryHandler)

	// Images
	// r.Get("/image/product/{source}/{id}", s.productImageHandler)

	// Watch
	r.Post("/watch/product/{id}", s.watchProductHandler)
	r.Post("/watch/verify/{hash}", s.watchVerifyHandler)
	r.Post("/watch/unsubscribe/{hash}", s.watchUnsubscribeHandler)

	// Go to
	r.Get("/goto/{id}", s.gotoHandler)

	// Misc
	r.Get("/search", s.searchHandler)
	r.Post("/contact", s.contactHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%s", s.Port), r)
	if err != nil {
		return err
	}

	return nil
}
