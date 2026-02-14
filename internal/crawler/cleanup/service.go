package cleanup

import (
	"go-wordpress/internal/crawler/domain"
	"regexp"
	"strconv"
	"strings"
)

func Cleans(cats []domain.Category) {
	for index, cat := range cats {
		cleaned := Clean(&cat)
		cats[index].CleanedProducts = cleaned
	}
}

func Clean(cat *domain.Category) []domain.ProductClean {

	cleaned := make([]domain.ProductClean, 0, len(cat.Products))

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

		// keep only digits
		reDigits := regexp.MustCompile(`\D`)
		priceStr = reDigits.ReplaceAllString(priceStr, "")

		priceStr = convertPersianDigits(priceStr)

		var priceInt int
		if priceStr != "" {
			if val, err := strconv.Atoi(priceStr); err == nil {
				priceInt = val
			}
		}

		// --- Build cleaned product ---
		cleaned = append(cleaned, domain.ProductClean{
			Title:       title,
			Price:       int64(priceInt),
			Link:        p.Link,
			Image:       p.Image,
			Description: p.Description,
		})
	}

	return cleaned
}

func convertPersianDigits(str string) string {
	persianDigits := []string{"۰", "۱", "۲", "۳", "۴", "۵", "۶", "۷", "۸", "۹"}
	for i, digit := range persianDigits {
		str = strings.ReplaceAll(str, digit, strconv.Itoa(i))
	}
	return str
}
