// Package extract runs the stateless pipeline for one product page:
// config, fetch, page_check, extract, normalize, validate.
package extract

import (
	"encoding/json"
	"time"

	"go-wordpress/internal/diagnose"
)

// Result statuses.
const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

// Price states.
const (
	PriceAvailable    = "available"
	PriceCallForPrice = "call_for_price"
	PriceOutOfStock   = "out_of_stock"
	PriceUnknown      = "unknown"
)

// Currency is always toman.
const CurrencyIRT = "IRT"

// Warning codes.
const (
	WarnFieldMissing = "FIELD_MISSING"
	WarnPriceMissing = "PRICE_MISSING"
)

// Request is the body of POST /api/v1/extract.
type Request struct {
	URL       string `json:"url"`
	RequestID string `json:"request_id"`
}

// Response is the single object returned for both results.
type Response struct {
	Status        string             `json:"status"`
	RequestID     string             `json:"request_id"`
	URL           string             `json:"url"`
	SiteID        string             `json:"site_id,omitempty"`
	ConfigVersion string             `json:"config_version,omitempty"`
	StartedAt     time.Time          `json:"started_at"`
	FinishedAt    time.Time          `json:"finished_at"`
	DurationMS    int64              `json:"duration_ms"`
	Product       *Product           `json:"product,omitempty"`
	Warnings      []Warning          `json:"warnings,omitempty"`
	Error         *diagnose.Failure  `json:"error,omitempty"`
	Evidence      *diagnose.Evidence `json:"evidence,omitempty"`
	Partial       *Partial           `json:"partial,omitempty"`
}

// MarshalJSON always writes "warnings" as an array on success, and omits
// it on failure.
func (r Response) MarshalJSON() ([]byte, error) {
	type alias Response
	w := struct {
		alias
		Warnings *[]Warning `json:"warnings,omitempty"`
	}{alias: alias(r)}
	if r.Status == StatusSuccess {
		ws := r.Warnings
		if ws == nil {
			ws = []Warning{}
		}
		w.Warnings = &ws
	}
	return json.Marshal(w)
}

// Product is the extracted product.
type Product struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	ImageURL    *string `json:"image_url"`
	Description *string `json:"description"`
	Price       Price   `json:"price"`
}

// Price holds prices in toman.
type Price struct {
	Current  *int64 `json:"current"`
	Regular  *int64 `json:"regular"`
	Currency string `json:"currency"`
	Raw      string `json:"raw"`
	State    string `json:"state"`
}

// Warning notes a non-fatal problem in a successful result.
type Warning struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Partial carries what was found when extraction failed.
type Partial struct {
	Product *Product `json:"product"`
}
