package domain

type Chunk struct {
	ID        string    `json:"id"`
	DocID     string    `json:"doc_id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Index     int       `json:"index"`
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding,omitempty"`
}
