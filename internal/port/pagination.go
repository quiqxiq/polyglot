package port

// PageFilter membatasi query daftar: tenant + limit/offset (F6-8).
// Zero value berarti tanpa filter (kompatibel FindAll).
type PageFilter struct {
	TenantID string
	Limit    int
	Offset   int
}
