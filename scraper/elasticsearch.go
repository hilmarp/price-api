package scraper

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/olivere/elastic/v7"
)

const searchIndex string = "price_search"

// Elasticsearch handles ES operations
type Elasticsearch struct {
	Client *elastic.Client
}

// CreateSearchIndex will only create the index if it doesn't exist
func (es *Elasticsearch) CreateSearchIndex() error {
	exists, err := es.Client.IndexExists(searchIndex).Do(context.TODO())
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		},
		"mappings": {
			"properties": {
				"ID": {
					"type": "keyword"
				},
				"ScrapedAt": {
					"type": "date"
				},
				"Source": {
					"type": "keyword"
				},
				"ProductCode": {
					"type": "text"
				},
				"Slug": {
					"type": "keyword"
				},
				"URL": {
					"type": "keyword"
				},
				"Title": {
					"type": "text"
				},
				"Categories": {
					"type": "keyword"
				},
				"Description": {
					"type": "text"
				},
				"MainImgURL": {
					"type": "keyword"
				},
				"Price": {
					"type": "long"
				},
				"OnSale": {
					"type": "boolean"
				}
			}
		}
	}`

	createIndex, err := es.Client.CreateIndex(searchIndex).BodyString(mapping).Do(context.TODO())
	if err != nil {
		return err
	}

	if !createIndex.Acknowledged {
		return fmt.Errorf("index not acknowledged")
	}

	return nil
}

// UpdateOrIndexSearchProduct add the product to an index or updates if it's not already there
func (es *Elasticsearch) UpdateOrIndexSearchProduct(product *SearchProduct) error {
	err := es.CreateSearchIndex()
	if err != nil {
		return err
	}

	_, err = es.Client.Index().
		Index(searchIndex).
		Id(strconv.Itoa(int(product.ID))).
		BodyJson(product).
		Do(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

// DeleteSearchProductByID deletes a product from ES
func (es *Elasticsearch) DeleteSearchProductByID(id uint) error {
	_, err := es.Client.Delete().
		Index(searchIndex).
		Id(strconv.Itoa(int(id))).
		Do(context.TODO())
	if err != nil {
		return fmt.Errorf("could not delete product id %v from ES: %w", id, err)
	}

	return nil
}

// SearchByURL returns products that match URL
func (es *Elasticsearch) SearchByURL(URL string) (*[]SearchProduct, error) {
	query := elastic.NewTermQuery("URL", URL)

	searchResult, err := es.Client.Search().
		Index(searchIndex).
		Query(query).
		Pretty(true).
		Do(context.TODO())
	if err != nil {
		return nil, err
	}

	var esProducts []SearchProduct
	var esProductType SearchProduct
	for _, item := range searchResult.Each(reflect.TypeOf(esProductType)) {
		p := item.(SearchProduct)
		esProducts = append(esProducts, p)
	}

	return &esProducts, nil
}

// SearchForProduct will return products from ES which match value
func (es *Elasticsearch) SearchForProduct(value string, limit, offset int) (*[]SearchProduct, error) {
	if formatters.IsValidURL(value) {
		products, err := es.SearchByURL(value)
		if err != nil {
			return nil, err
		}

		return products, nil
	}

	query := elastic.NewMultiMatchQuery(value, "ProductCode", "Title", "Description")

	searchResult, err := es.Client.Search().
		Index(searchIndex).
		Query(query).
		From(offset).
		Size(limit).
		Pretty(true).
		Do(context.TODO())
	if err != nil {
		return nil, err
	}

	var esProducts []SearchProduct
	var esProductType SearchProduct
	for _, item := range searchResult.Each(reflect.TypeOf(esProductType)) {
		p := item.(SearchProduct)
		esProducts = append(esProducts, p)
	}

	return &esProducts, nil
}
