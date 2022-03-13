package scraper

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"gorm.io/gorm"
)

// SQL handles SQL database operations
type SQL struct {
	*gorm.DB
}

// GetProducts returns a limit of products
func (db *SQL) GetProducts(limit, offset, priceFrom, priceTo int, order, onSale string, sources, categorySlugs []string) (*[]ProductPriceDiff, error) {
	sql, args := db.GetProductsSQLStmt(limit, offset, priceFrom, priceTo, order, onSale, sources, categorySlugs, false)

	// Get products in category with slug from slugs list
	var products []ProductPriceDiff
	result := db.Raw(sql, args...).Scan(&products)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &products, nil
}

// GetProductsCount returns the total count of products with filters
func (db *SQL) GetProductsCount(limit, offset, priceFrom, priceTo int, order, onSale string, sources, categorySlugs []string) (int, error) {
	sql, args := db.GetProductsSQLStmt(limit, offset, priceFrom, priceTo, order, onSale, sources, categorySlugs, true)

	// Get products in category with slug from slugs list
	type countJSON struct {
		Count int
	}
	var count countJSON
	result := db.Raw(sql, args...).Scan(&count)
	if err := result.Error; err != nil {
		return 0, err
	}

	return count.Count, nil
}

// GetProductsSQLStmt returns the SQL statement with filters to get products from DB
func (db *SQL) GetProductsSQLStmt(limit, offset, priceFrom, priceTo int, order, onSale string, sources, categorySlugs []string, countOnly bool) (string, []interface{}) {
	// Only add parent categories to list, all children will be returned
	// Query might have heyrnartol]hljod-og-mynd and hljod-og-mynd, so only add hljod-og-mynd
	var slugs []string
	for _, s := range categorySlugs {
		sSlugs := strings.Split(s, "]")

		// If no separator it's a parent, so add it
		if len(sSlugs) < 2 {
			slugs = append(slugs, s)
			continue
		}

		// Check for parent, if it contains any of the parents, don't add it
		// Might be heyrnartol]hljod-og-mynd and tolvuheyrnartol]heyrnartol]hljod-og-mynd, so only add the first one
		var uc UniqueCategory
		db.Where("slug = ?", s).First(&uc)

		if formatters.IsInStringList(uc.Parent, categorySlugs) {
			continue
		}

		slugs = append(slugs, s)
	}

	// Set up statements and arguments
	var args []interface{}

	// Slugs where
	slugWheres := make([]string, len(slugs))
	for i, s := range slugs {
		args = append(args, s)
		slugWheres[i] = "c.slug = ?"
	}
	slugStmt := ""
	if len(slugWheres) > 0 {
		slugStmt = fmt.Sprintf("(%s)", strings.Join(slugWheres, " OR "))
	}

	// Sources where
	sourceWheres := make([]string, len(sources))
	for i, s := range sources {
		args = append(args, s)
		sourceWheres[i] = "p.source = ?"
	}
	sourceStmt := ""
	if len(sourceWheres) > 0 {
		sourceStmt = fmt.Sprintf("(%s)", strings.Join(sourceWheres, " OR "))
	}

	// Price where
	priceStmt := ""
	if priceFrom > 0 || priceTo > 0 {
		if priceFrom > 0 && priceTo > 0 {
			args = append(args, priceFrom)
			args = append(args, priceTo)
			priceStmt = "(price >= ? AND price <= ?)"
		} else if priceFrom > 0 {
			args = append(args, priceFrom)
			priceStmt = "(price >= ?)"
		} else if priceTo > 0 {
			args = append(args, priceTo)
			priceStmt = "(price <= ?)"
		}
	}

	// OnSale where
	onSaleStmt := ""
	if onSale != "" {
		onSaleVal := "1"
		if onSale == "false" {
			onSaleVal = "0"
		}
		args = append(args, onSaleVal)
		onSaleStmt = "(on_sale = ?)"
	}

	// All where clauses combined
	whereStmt := ""
	whereStmts := []string{slugStmt, sourceStmt, priceStmt, onSaleStmt}
	whereStmtsNotEmpty := make([]string, 0)
	for _, s := range whereStmts {
		if s != "" {
			whereStmtsNotEmpty = append(whereStmtsNotEmpty, s)
		}
	}
	whereStmt = strings.Join(whereStmtsNotEmpty, " AND ")
	if whereStmt != "" {
		whereStmt = fmt.Sprintf("WHERE %s", whereStmt)
	}

	// Order by
	orderByStmt := ""
	if order != "" {
		orderByStmt = fmt.Sprintf("ORDER BY %s", order)
	}

	// Limit offset args
	if !countOnly {
		args = append(args, limit)
		args = append(args, offset)
	}

	sql := ""
	if countOnly {
		if len(slugWheres) > 0 {
			sql = fmt.Sprintf(
				`
					SELECT count(*) AS count
					FROM products AS p
					INNER JOIN categories AS c
					ON p.id = c.product_id
					%s
				`,
				whereStmt,
			)
		} else {
			sql = fmt.Sprintf(
				`
					SELECT count(*) AS count
					FROM products AS p
					%s
				`,
				whereStmt,
			)
		}
	} else {
		if len(slugWheres) > 0 {
			sql = fmt.Sprintf(
				`
					SELECT p.*, ppc.price_diff, ppc.price_lower FROM products AS p
					INNER JOIN categories AS c
					ON p.id = c.product_id
					LEFT JOIN product_price_changes AS ppc
					ON p.id = ppc.product_id
					%s
					%s
					LIMIT ? OFFSET ?
				`,
				whereStmt,
				orderByStmt,
			)
		} else {
			sql = fmt.Sprintf(
				`
					SELECT p.*, ppc.price_diff, ppc.price_lower FROM products AS p
					LEFT JOIN product_price_changes AS ppc
					ON p.id = ppc.product_id
					%s
					%s
					LIMIT ? OFFSET ?
				`,
				whereStmt,
				orderByStmt,
			)
		}
	}

	return sql, args
}

// GetUniqueCategories returns unique categories
func (db *SQL) GetUniqueCategories(parent string) (*[]UniqueCategory, error) {
	var categories []UniqueCategory
	result := db.Where("parent = ?", parent).
		Order("name asc").
		Find(&categories)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &categories, nil
}

// GetUniqueCategoryBySlug returns a single category by slug
func (db *SQL) GetUniqueCategoryBySlug(slug string) (*UniqueCategory, error) {
	var category UniqueCategory
	result := db.Where("slug = ?", slug).First(&category)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &category, nil
}

// GetProductsByIDs returns products with ID
func (db *SQL) GetProductsByIDs(ids ...uint) (*[]Product, error) {
	var products []Product

	if len(ids) == 0 {
		return &products, nil
	}

	idParamsS := make([]string, len(ids))
	for i := range ids {
		idParamsS[i] = "?"
	}

	// SELECT * FROM table WHERE id IN (5,4,3,1,6) ORDER BY FIELD(id, 5,4,3,1,6);

	idParams := strings.Join(idParamsS, ",")
	sql := fmt.Sprintf(`
		SELECT * FROM products
		WHERE id IN (%s)
		ORDER BY FIELD(id, %s)
	`, idParams, idParams)

	// Ex. 1,2,3,1,2,3 because we pass them in twice as parameters
	// TODO: One loop
	var doubleIds []uint
	doubleIds = append(doubleIds, ids...)
	doubleIds = append(doubleIds, ids...)

	args := make([]interface{}, len(doubleIds))
	for i, id := range doubleIds {
		args[i] = id
	}

	result := db.Raw(sql, args...).Scan(&products)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &products, nil
}

// GetProductsPriceChanges returns products that have changed in price recently,
// both price drop and price hikes
func (db *SQL) GetProductsPriceChanges(limit, offset int, lower string) (*[]ProductPriceDiff, error) {
	whereLower := ""
	switch lower {
	case "true":
		whereLower = "AND ppc.price_lower = 1"
	case "false":
		whereLower = "AND ppc.price_lower = 0"
	}

	sql := fmt.Sprintf(`
		SELECT p.*, ppc.price_diff, ppc.price_lower FROM products AS p
		INNER JOIN product_price_changes AS ppc
		ON p.id = ppc.product_id
		WHERE ppc.price_diff > 0
		%s
		ORDER BY ppc.price_diff DESC
		LIMIT ? OFFSET ?
	`, whereLower)

	var products []ProductPriceDiff
	result := db.Raw(sql, limit, offset).Scan(&products)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &products, nil
}

// GetProductByID returns a single product by ID
func (db *SQL) GetProductByID(id uint) (*Product, error) {
	var product Product
	result := db.First(&product, id)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// DeleteProductByID will delete everything related to that product id from all tables
func (db *SQL) DeleteProductByID(id uint) error {
	result := db.Where("product_id = ?", id).Unscoped().Delete(Price{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("product_id = ?", id).Unscoped().Delete(Image{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("product_id = ?", id).Unscoped().Delete(Stock{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("product_id = ?", id).Unscoped().Delete(Spec{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("product_id = ?", id).Unscoped().Delete(Category{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("product_id = ?", id).Unscoped().Delete(WatchProduct{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("product_id = ?", id).Unscoped().Delete(ProductViewCount{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("product_id = ?", id).Unscoped().Delete(ProductPriceChange{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("product_id = ?", id).Unscoped().Delete(ProductClickCount{})
	if err := result.Error; err != nil {
		return err
	}

	result = db.Where("id = ?", id).Unscoped().Delete(Product{})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// GetProductByURL returns a single product by URL
func (db *SQL) GetProductByURL(url string) (*Product, error) {
	var product Product
	result := db.Where("url = ?", url).First(&product)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// GetProductBySlug returns a single product
func (db *SQL) GetProductBySlug(slug string) (*Product, error) {
	var product Product
	result := db.Where("slug = ?", slug).First(&product)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// GetProductsBySourceProductCode returns products with source (elko.is, ...) and product code
func (db *SQL) GetProductsBySourceProductCode(source, productCode string) (*[]Product, error) {
	sql := `
		SELECT * FROM products
		WHERE source = ? AND product_code = ?
		ORDER BY url DESC
	`

	var foundProducts []Product
	result := db.Raw(sql, source, productCode).Scan(&foundProducts)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &foundProducts, nil
}

// GetProductsBySourceProductCodeTitle returns products with source (elko.is, ...), product code and title
func (db *SQL) GetProductsBySourceProductCodeTitle(source, productCode, title string) (*[]Product, error) {
	sql := `
		SELECT * FROM products
		WHERE source = ? AND product_code = ? AND title = ?
		ORDER BY url DESC
	`

	var foundProducts []Product
	result := db.Raw(sql, source, productCode, title).Scan(&foundProducts)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &foundProducts, nil
}

// GetProductPrices returns prices for a product
func (db *SQL) GetProductPrices(id uint, from time.Time, order string) (*[]Price, error) {
	sql := fmt.Sprintf(`
		SELECT * FROM prices
		WHERE product_id = ?
		AND date >= ?
		ORDER BY %s;
	`, order)

	var prices []Price
	result := db.Raw(sql, id, from).Scan(&prices)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &prices, nil
}

// GetProductSpecs returns specs for a product
func (db *SQL) GetProductSpecs(id uint) (*[]Spec, error) {
	var specs []Spec
	result := db.Where("product_id = ?", id).Find(&specs)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &specs, nil
}

// GetProductStocks returns stocks for a product
func (db *SQL) GetProductStocks(id uint) (*[]Stock, error) {
	var stocks []Stock
	result := db.Where("product_id = ?", id).Find(&stocks)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &stocks, nil
}

// GetProductImages returns images for a product
func (db *SQL) GetProductImages(id uint) (*[]Image, error) {
	var images []Image
	result := db.Where("product_id = ?", id).Find(&images)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &images, nil
}

// GetProductCategories returns categories for a product
func (db *SQL) GetProductCategories(id uint) (*[]Category, error) {
	var categories []Category
	result := db.Where("product_id = ?", id).Find(&categories)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &categories, nil
}

// InsertUniqueCategory will insert a category in the unique_category table,
// but only if it doesn't exists, so no duplicate categories
func (db *SQL) InsertUniqueCategory(categories *[]Category) error {
	for _, category := range *categories {
		var uniqueCategories []UniqueCategory
		result := db.Where("name = ? AND slug = ? AND parent = ?", category.Name, category.Slug, category.Parent).
			Find(&uniqueCategories)
		if err := result.Error; err != nil {
			return err
		}

		// Insert if not exists
		if len(uniqueCategories) == 0 {
			uc := UniqueCategory{
				Name:   category.Name,
				Slug:   category.Slug,
				Parent: category.Parent,
			}

			result := db.Create(&uc)
			if err := result.Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// UpdateOrCreateProduct will update product fields if it already exists,
// otherwise create it, and return the db instance of product
func (db *SQL) UpdateOrCreateProduct(scrapedProduct *Product) (*Product, error) {
	// Insert unique categories
	err := db.InsertUniqueCategory(&scrapedProduct.Categories)
	if err != nil {
		return nil, err
	}

	// Check if it's been scraped already
	foundProduct, err := db.GetProductByURL(scrapedProduct.URL)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// Create if no product was found
	if foundProduct == nil {
		result := db.Create(scrapedProduct)
		if err := result.Error; err != nil {
			return nil, fmt.Errorf("error creating %v: %w", scrapedProduct.URL, err)
		}
		return scrapedProduct, nil // skip rest
	}

	result := db.Model(&foundProduct).Updates(map[string]interface{}{
		"source":       scrapedProduct.Source,
		"product_code": scrapedProduct.ProductCode,
		"slug":         scrapedProduct.Slug,
		"url":          scrapedProduct.URL,
		"title":        scrapedProduct.Title,
		"description":  scrapedProduct.Description,
		"main_img_url": scrapedProduct.MainImgURL,
		"price":        scrapedProduct.Price,
		"on_sale":      scrapedProduct.OnSale,
	})
	if err := result.Error; err != nil {
		return nil, err
	}

	// Delete the old specs and add the newly scraped
	db.Where("product_id = ?", foundProduct.ID).Unscoped().Delete(Spec{})

	for _, s := range scrapedProduct.Specs {
		db.Create(&Spec{
			Key:       s.Key,
			Value:     s.Value,
			ProductID: foundProduct.ID,
		})
	}

	// Delete old stock info and add new
	db.Where("product_id = ?", foundProduct.ID).Unscoped().Delete(Stock{})

	for _, s := range scrapedProduct.Stocks {
		db.Create(&Stock{
			Location:  s.Location,
			InStock:   s.InStock,
			ProductID: foundProduct.ID,
		})
	}

	// Delete the old image urls and add the newly scraped
	db.Where("product_id = ?", foundProduct.ID).Unscoped().Delete(Image{})

	for _, img := range scrapedProduct.AllImgURLs {
		db.Create(&Image{
			URL:         img.URL,
			OriginalURL: img.OriginalURL,
			ProductID:   foundProduct.ID,
		})
	}

	// Add new dated price
	if len(scrapedProduct.Prices) > 0 {
		price := Price{
			Price:     scrapedProduct.Prices[0].Price,
			Date:      scrapedProduct.Prices[0].Date,
			ProductID: foundProduct.ID,
		}
		result := db.Create(&price)
		if err := result.Error; err != nil {
			return nil, err
		}
	}

	// Delete the old categories and add the newly scraped
	db.Where("product_id = ?", foundProduct.ID).Unscoped().Delete(Category{})

	for _, c := range scrapedProduct.Categories {
		result := db.Create(&Category{
			Name:      c.Name,
			Slug:      c.Slug,
			Parent:    c.Parent,
			ProductID: foundProduct.ID,
		})
		if err := result.Error; err != nil {
			return nil, fmt.Errorf("error storing product categories: %w", err)
		}
	}

	return foundProduct, nil
}

// CreateWatchProduct will create a WatchProduct entry, email watching a product
func (db *SQL) CreateWatchProduct(watchProduct *WatchProduct) error {
	result := db.Create(watchProduct)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// UpdateWatchProduct will update a watchProduct with ID
func (db *SQL) UpdateWatchProduct(watchProduct *WatchProduct) error {
	if watchProduct.ID == 0 {
		return fmt.Errorf("no watchProduct ID")
	}

	result := db.Model(watchProduct).Updates(map[string]interface{}{
		"email":            watchProduct.Email,
		"product_id":       watchProduct.ProductID,
		"sent":             watchProduct.Sent,
		"price_id_sent":    watchProduct.PriceIDSent,
		"verified":         watchProduct.Verified,
		"verify_hash":      watchProduct.VerifyHash,
		"unsubscribe_hash": watchProduct.UnsubscribeHash,
	})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// GetWatchProducts returns a limit list of verified watches
func (db *SQL) GetWatchProducts(limit, offset int) (*[]WatchProduct, error) {
	var watchProducts []WatchProduct

	result := db.
		Where("verified = ?", true).
		Limit(limit).
		Offset(offset).
		Find(&watchProducts)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &watchProducts, nil
}

// GetWatchProductsByProductID returns watches for a product_id
func (db *SQL) GetWatchProductsByProductID(id uint) (*[]WatchProduct, error) {
	var watchProducts []WatchProduct

	result := db.
		Where("product_id = ?", id).
		Find(&watchProducts)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &watchProducts, nil
}

// GetWatchProductByVerifyHash returns a single WatchProduct by verify hash
func (db *SQL) GetWatchProductByVerifyHash(verifyHash string) (*WatchProduct, error) {
	var watchProduct WatchProduct
	result := db.Where("verify_hash = ?", verifyHash).First(&watchProduct)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &watchProduct, nil
}

// GetWatchProductByUnsubscribeHash returns a single WatchProduct by unsubscribe hash
func (db *SQL) GetWatchProductByUnsubscribeHash(unsubscribeHash string) (*WatchProduct, error) {
	var watchProduct WatchProduct
	result := db.Where("unsubscribe_hash = ?", unsubscribeHash).First(&watchProduct)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &watchProduct, nil
}

// DeleteWatchProductByUnsubscribeHash deletes a watch, so unsubscribe
func (db *SQL) DeleteWatchProductByUnsubscribeHash(unsubscribeHash string) error {
	result := db.Where("unsubscribe_hash = ?", unsubscribeHash).Unscoped().Delete(WatchProduct{})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// DeleteWatchProductByProductID deletes a watch by product id
func (db *SQL) DeleteWatchProductByProductID(id uint) error {
	result := db.Where("product_id = ?", id).Unscoped().Delete(WatchProduct{})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// GetProductViewCounts returns a limit list of view counts
func (db *SQL) GetProductViewCounts(limit, offset int) (*[]ProductViewCount, error) {
	var productViewCounts []ProductViewCount

	result := db.
		Limit(limit).
		Offset(offset).
		Find(&productViewCounts)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &productViewCounts, nil
}

// GetProductViewCount returns a product view count
func (db *SQL) GetProductViewCount(id uint) (*ProductViewCount, error) {
	var productViewCount ProductViewCount
	result := db.Where("product_id = ?", id).First(&productViewCount)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &productViewCount, nil
}

// IncrementProductViewCount increments a view count by 1
func (db *SQL) IncrementProductViewCount(id uint) error {
	foundProductViewCount, err := db.GetProductViewCount(id)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	// Doesn't exist, create it with value of 1
	if foundProductViewCount == nil {
		productViewCount := ProductViewCount{
			ProductID: id,
			Views:     1,
		}
		result := db.Create(&productViewCount)
		if err := result.Error; err != nil {
			return err
		}
		return nil // skip rest
	}

	// Update
	foundProductViewCount.Views = foundProductViewCount.Views + 1
	result := db.Save(foundProductViewCount)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// SetProductViewCount sets a view count to a specific value
func (db *SQL) SetProductViewCount(id uint, val int) error {
	foundProductViewCount, err := db.GetProductViewCount(id)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	// Doesn't exist, create it with value
	if foundProductViewCount == nil {
		productViewCount := ProductViewCount{
			ProductID: id,
			Views:     val,
		}
		result := db.Create(&productViewCount)
		if err := result.Error; err != nil {
			return err
		}
		return nil // skip rest
	}

	// Update
	foundProductViewCount.Views = val
	result := db.Save(foundProductViewCount)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// DeleteProductViewCountByID deletes a view counter by ID
func (db *SQL) DeleteProductViewCountByID(id uint) error {
	result := db.Where("id = ?", id).Unscoped().Delete(ProductViewCount{})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// GetPopularProducts returns the most popular products, based on product view count
func (db *SQL) GetPopularProducts(limit, offset int) (*[]ProductPriceDiff, error) {
	sql := `
		SELECT p.*, ppc.price_diff, ppc.price_lower FROM products AS p
		INNER JOIN product_view_counts AS pvc
		ON p.id = pvc.product_id
		LEFT JOIN product_price_changes AS ppc
		ON p.id = ppc.product_id
		WHERE pvc.views > 0
		ORDER BY pvc.views DESC
		LIMIT ? OFFSET ?
	`

	var products []ProductPriceDiff
	result := db.Raw(sql, limit, offset).Scan(&products)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &products, nil
}

// GetProductPriceChangeByProductID returns a product price change entry
func (db *SQL) GetProductPriceChangeByProductID(productID uint) (*ProductPriceChange, error) {
	var product ProductPriceChange
	result := db.Where("product_id = ?", productID).First(&product)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// UpdateOrCreateProductPriceChange will create/update a product price entry
func (db *SQL) UpdateOrCreateProductPriceChange(productPriceChange *ProductPriceChange) error {
	found, err := db.GetProductPriceChangeByProductID(productPriceChange.ProductID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	if found == nil {
		return db.CreateProductPriceChange(productPriceChange)
	}

	// Update
	productPriceChange.ID = found.ID
	return db.UpdateProductPriceChange(productPriceChange)
}

// UpdateProductPriceChange will update a product price entry
func (db *SQL) UpdateProductPriceChange(productPriceChange *ProductPriceChange) error {
	if productPriceChange.ID == 0 {
		return fmt.Errorf("no productPriceChange ID")
	}

	result := db.Model(productPriceChange).Updates(map[string]interface{}{
		"product_id":      productPriceChange.ProductID,
		"price_diff":      productPriceChange.PriceDiff,
		"price_lower":     productPriceChange.PriceLower,
		"prev_price_date": productPriceChange.PrevPriceDate,
	})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// CreateProductPriceChange will create a product price entry
func (db *SQL) CreateProductPriceChange(productPriceChange *ProductPriceChange) error {
	result := db.Create(productPriceChange)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// DeleteProductPriceChangeByID deletes a view counter by ID
func (db *SQL) DeleteProductPriceChangeByID(id uint) error {
	result := db.Where("id = ?", id).Unscoped().Delete(ProductPriceChange{})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// CreateProductClickCount creates a product click count db entry
func (db *SQL) CreateProductClickCount(productClickCount *ProductClickCount) error {
	result := db.Create(productClickCount)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// GetLastUpdatedProduct returns the last updated product
func (db *SQL) GetLastUpdatedProduct() (*Product, error) {
	var product Product
	result := db.
		Order("updated_at desc").
		Limit(1).
		First(&product)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// GetBotByURL returns a single bot by URL
func (db *SQL) GetBotByURL(URL string) (*Bot, error) {
	var bot Bot
	result := db.Where("url = ?", URL).First(&bot)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &bot, nil
}

// UpdateOrCreateBot updates or creates a bot
func (db *SQL) UpdateOrCreateBot(bot *Bot) error {
	foundBot, err := db.GetBotByURL(bot.URL)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	// Create
	if foundBot == nil {
		result := db.Create(bot)
		if err := result.Error; err != nil {
			return err
		}
		return nil // skip rest
	}

	// Update
	result := db.Model(foundBot).Updates(map[string]interface{}{
		"url":         bot.URL,
		"started_at":  bot.StartedAt,
		"finished_at": bot.FinishedAt,
	})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// Migrate will run db migration
func (db *SQL) Migrate() error {
	err := db.AutoMigrate(
		&Product{},
		&Price{},
		&Image{},
		&Stock{},
		&Spec{},
		&Category{},
		&UniqueCategory{},
		&WatchProduct{},
		&ProductViewCount{},
		&ProductPriceChange{},
		&ProductClickCount{},
		&Bot{},
	)
	if err != nil {
		return err
	}

	return nil
}
