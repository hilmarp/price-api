package scraper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"bitbucket.org/hilmarp/price-scraper/formatters"
)

func saveAndSetProductImages(db *SQL, scrapedProduct *Product) error {
	if len(scrapedProduct.AllImgURLs) > 0 {
		scrapedProduct.MainImgURL = scrapedProduct.AllImgURLs[0].URL
	}

	foundProducts, err := db.GetProductsBySourceProductCode(scrapedProduct.Source, scrapedProduct.ProductCode)
	if err != nil {
		return err
	}

	// New product, just save images without checking anything
	if len(*foundProducts) == 0 {
		for i := range scrapedProduct.AllImgURLs {
			path, err := saveProductImgFromURL(scrapedProduct.AllImgURLs[i].OriginalURL, scrapedProduct.Source)
			if err != nil {
				return err
			}
			scrapedProduct.AllImgURLs[i].URL = fmt.Sprintf("https://api.verdfra.is/image/product/%v/%v", scrapedProduct.Source, path)
		}

		if len(scrapedProduct.AllImgURLs) > 0 {
			scrapedProduct.MainImgURL = scrapedProduct.AllImgURLs[0].URL
		}

		return nil // skip rest
	}

	// Update, check if images have already been saved, then use that product onward
	var foundProduct *Product
	for _, product := range *foundProducts {
		for _, img := range product.AllImgURLs {
			if strings.Contains(img.URL, "https://api.verdfra.is/image/product/") {
				foundProduct = &product
				break
			}
		}
	}

	// Found a product with saved images to use, now check if they need updating,
	// if so, delete the old one and save a new one
	if foundProduct != nil {
		needUpdate := false
		if len(foundProduct.AllImgURLs) != len(scrapedProduct.AllImgURLs) {
			needUpdate = true
		}

		if len(foundProduct.AllImgURLs) == len(scrapedProduct.AllImgURLs) {
			for i := range foundProduct.AllImgURLs {
				if foundProduct.AllImgURLs[i].OriginalURL != scrapedProduct.AllImgURLs[i].OriginalURL {
					needUpdate = true
					break
				}
			}
		}

		if needUpdate {
			// Delete old ones
			for _, img := range foundProduct.AllImgURLs {
				err := deleteProductImg(strings.ReplaceAll(img.URL, "https://api.verdfra.is/image/product/"+scrapedProduct.Source+"/", ""), scrapedProduct.Source)
				if err != nil {
					return err
				}
			}

			// Save new ones
			for i := range scrapedProduct.AllImgURLs {
				path, err := saveProductImgFromURL(scrapedProduct.AllImgURLs[i].OriginalURL, scrapedProduct.Source)
				if err != nil {
					return err
				}
				scrapedProduct.AllImgURLs[i].URL = fmt.Sprintf("https://api.verdfra.is/image/product/%v/%v", scrapedProduct.Source, path)
			}
		}

		if !needUpdate {
			// Use the same ones as in the found product
			for i := range scrapedProduct.AllImgURLs {
				scrapedProduct.AllImgURLs[i].URL = foundProduct.AllImgURLs[i].URL
			}
		}

		if len(scrapedProduct.AllImgURLs) > 0 {
			scrapedProduct.MainImgURL = scrapedProduct.AllImgURLs[0].URL
		}

		return nil // skip rest
	}

	// Didn't find a product with saved images, so save them
	for i := range scrapedProduct.AllImgURLs {
		path, err := saveProductImgFromURL(scrapedProduct.AllImgURLs[i].OriginalURL, scrapedProduct.Source)
		if err != nil {
			return err
		}
		scrapedProduct.AllImgURLs[i].URL = fmt.Sprintf("https://api.verdfra.is/image/product/%v/%v", scrapedProduct.Source, path)
	}

	if len(scrapedProduct.AllImgURLs) > 0 {
		scrapedProduct.MainImgURL = scrapedProduct.AllImgURLs[0].URL
	}

	return nil
}

// saveProductImgFromURL saves the product image to a static folder
// and returns the path, the path is a random string
func saveProductImgFromURL(url, source string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	absPath := os.Getenv("PRICE_ABS_PATH")
	fileName := formatters.GetRandomStringWithTimestamp(15)
	file, err := os.Create(fmt.Sprintf("%s/static/img/products/%s/%s.jpg", absPath, source, fileName))
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

// deleteProductImg deletes the product image at path
func deleteProductImg(path, source string) error {
	absPath := os.Getenv("PRICE_ABS_PATH")

	fullPath := fmt.Sprintf("%s/static/img/products/%s/%s.jpg", absPath, source, path)
	err := os.Remove(fullPath)
	if err != nil {
		return err
	}

	return nil
}
