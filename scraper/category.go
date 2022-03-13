package scraper

import (
	"strings"

	"bitbucket.org/hilmarp/price-scraper/formatters"
	"github.com/PuerkitoBio/goquery"
)

func getCategoriesFromArray(breadcrumbs []string) []Category {
	categories := make([]Category, 0)
	breadcrumbsMap := make(map[int]string)

	for i, bc := range breadcrumbs {
		text := bc

		// Add to map so we can check parents
		breadcrumbsMap[i] = text

		// Check all previous breadcrumbs, those are the parents of this one
		var parents []string
		j := i - 1
		for {
			val, ok := breadcrumbsMap[j]
			if !ok {
				break
			}

			parents = append(parents, val)
			j = j - 1
		}

		parentsStr := strings.Join(parents, "]")

		// Want the slug to include parent categories
		slug := formatters.GetSlug(text)
		parentSlug := ""

		if parentsStr != "" {
			// Want to keep the ] char
			var slugParents []string
			for _, p := range parents {
				slugParents = append(slugParents, formatters.GetSlug(p))
			}

			slug = slug + "]" + strings.Join(slugParents, "]")
			parentSlug = strings.Join(slugParents, "]")
		}

		category := Category{
			Name:   text,
			Slug:   slug,
			Parent: parentSlug,
		}
		categories = append(categories, category)
	}

	return categories
}

func getCategoriesFromBreadcrumbs(breadcrumbs *goquery.Selection, firstItem, lastItem bool) []Category {
	categories := make([]Category, 0)
	breadcrumbsMap := make(map[int]string)
	breadcrumbs.Each(func(i int, s *goquery.Selection) {
		// Which items to check, default all
		check := true

		if !firstItem && !lastItem {
			// We don't want the first one, or the last one since it's just the product title
			check = i != 0 && i != breadcrumbs.Length()-1
		}

		if firstItem && !lastItem {
			// We don't want the last one since it's just the product title
			check = i != breadcrumbs.Length()-1
		}

		if !firstItem && lastItem {
			// We don't want the first one, since it's just a link to the frontpage
			check = i != 0
		}

		if check {
			text := strings.TrimSpace(s.Text())

			// Add to map so we can check parents
			breadcrumbsMap[i] = text

			// Check all previous breadcrumbs, those are the parents of this one
			var parents []string
			j := i - 1
			for {
				val, ok := breadcrumbsMap[j]
				if !ok {
					break
				}

				parents = append(parents, val)
				j = j - 1
			}

			parentsStr := strings.Join(parents, "]")

			// Want the slug to include parent categories
			slug := formatters.GetSlug(text)
			parentSlug := ""

			if parentsStr != "" {
				// Want to keep the ] char
				var slugParents []string
				for _, p := range parents {
					slugParents = append(slugParents, formatters.GetSlug(p))
				}

				slug = slug + "]" + strings.Join(slugParents, "]")
				parentSlug = strings.Join(slugParents, "]")
			}

			category := Category{
				Name:   text,
				Slug:   slug,
				Parent: parentSlug,
			}
			categories = append(categories, category)
		}
	})

	return categories
}
