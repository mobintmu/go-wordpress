package shared

type Pagination struct {
	Limit  int64 `json:"limit" form:"limit"`
	Offset int64 `json:"offset" form:"offset"`
}
