package formatters

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

// IsValidURL checks if a string is a valid URL
func IsValidURL(URL string) bool {
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return false
	}

	u, err := url.Parse(URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// GetURLHost returns host of URL, without www, ex. www.tl.is => tl.is
func GetURLHost(URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return ""
	}

	host := u.Host
	if strings.HasPrefix(host, "www.") {
		return strings.Replace(host, "www.", "", 1)
	}
	return host
}

//GetURLWithoutWWW returns the URL without the www. prefix
func GetURLWithoutWWW(URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return URL
	}

	if strings.HasPrefix(u.Host, "www.") {
		URL = strings.Replace(URL, "www.", "", 1)
	}

	return URL
}

//GetURLWithWWW returns the URL with the www. prefix
func GetURLWithWWW(URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return URL
	}

	if strings.HasPrefix(u.Host, "www.") {
		return URL
	}

	if u.Scheme == "https" {
		URL = strings.Replace(URL, "https://", "https://www.", 1)
	} else {
		URL = strings.Replace(URL, "http://", "http://www.", 1)
	}

	return URL
}

// GetURLWithoutQuery returns URL without query params
// ex. https://heimkaup.is/nuby-gomlaga-snud-glow?vid=28743 => https://heimkaup.is/nuby-gomlaga-snud-glow
func GetURLWithoutQuery(URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return URL
	}

	return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
}

// GetURLWithQueryParam returns URL with added query param
func GetURLWithQueryParam(URL, key, value string) string {
	u, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return URL
	}

	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()

	return u.String()
}

// GetCleanURL returns URL without query params, unless they're in the list of ones to keep
func GetCleanURL(URL string, queryParamsToKeep []string) string {
	u, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return URL
	}

	cleanURL := GetURLWithoutQuery(URL)

	// Add query params to keep back on to URL
	params := u.Query()
	for _, param := range queryParamsToKeep {
		if val, ok := params[param]; ok && len(val) > 0 {
			cleanURL = GetURLWithQueryParam(cleanURL, param, val[0])
		}
	}

	return cleanURL
}
