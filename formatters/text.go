package formatters

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

func StringToPrice(s string) uint {
	price := strings.TrimSpace(s)
	price = strings.ReplaceAll(price, "\u00A0", " ") // nbsp
	price = strings.Split(price, " ")[0]
	price = strings.ReplaceAll(price, ".", "")

	priceInt := StringToInt(price)

	return uint(priceInt)
}

func StringToInt(s string) int {
	sInt, err := strconv.Atoi(s)
	if err != nil {
		sInt = 0
	}

	return sInt
}

// GetSlug will create a URL slug from the strings passed in
func GetSlug(texts ...string) string {
	var slugs []string
	for _, text := range texts {
		slugs = append(slugs, slug.Make(text))
	}

	return strings.Join(slugs, "-")
}

// GetRandomNumString returns a random number string
// between low and hi numbers, ex. 132457
func GetRandomNumString(low, hi int) string {
	num := low + rand.Intn(hi-low)

	return strconv.Itoa(num)
}

// GetRandomString returns a random string of length n
func GetRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// GetRandomStringWithTimestamp returns a random string of length n,
// with prepended timestamp
func GetRandomStringWithTimestamp(n int) string {
	now := time.Now().UnixNano()

	return strconv.FormatInt(now, 10) + GetRandomString(n)
}
