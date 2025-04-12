package productDto

type CreateProductRequest struct {
	Type  string `json:"type"`
	PvzID string `json:"pvzId"`
}
