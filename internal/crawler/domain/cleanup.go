package domain

import (
	"regexp"
	"strconv"
	"strings"
)

func (web *Website) Cleans() {
	for index := range web.Categories {
		web.Categories[index].Clean()
	}
}

func (cat *Category) Clean() {

	cleaned := make([]ProductClean, 0, len(cat.Products))

	// regex to strip HTML tags
	reTags := regexp.MustCompile(`<[^>]*>`)

	for _, p := range cat.Products {
		// --- Clean Title ---
		title := reTags.ReplaceAllString(p.Title, "") // remove tags
		title = strings.TrimSpace(title)

		// --- Clean Price ---
		// Example: "499,000 تومان"
		priceStr := p.Price
		// remove currency word
		priceStr = strings.ReplaceAll(priceStr, "تومان", "")
		// remove commas and non-digits
		priceStr = strings.ReplaceAll(priceStr, ",", "")
		priceStr = strings.TrimSpace(priceStr)

		// Convert first: \D is ASCII-only and would delete Persian digits.
		priceStr = convertPersianDigits(priceStr)

		// keep only digits
		reDigits := regexp.MustCompile(`\D`)
		priceStr = reDigits.ReplaceAllString(priceStr, "")

		var priceInt int
		if priceStr != "" {
			if val, err := strconv.Atoi(priceStr); err == nil {
				priceInt = val
			}
		}

		// --- Build cleaned product ---
		cleaned = append(cleaned, ProductClean{
			Title:       title,
			Price:       int64(priceInt),
			Link:        p.Link,
			Image:       p.Image,
			Description: p.Description,
		})
	}

	cat.CleanedProducts = cleaned
}

func convertPersianDigits(str string) string {
	persianDigits := []string{"۰", "۱", "۲", "۳", "۴", "۵", "۶", "۷", "۸", "۹"}
	for i, digit := range persianDigits {
		str = strings.ReplaceAll(str, digit, strconv.Itoa(i))
	}
	return str
}
