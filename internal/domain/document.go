package domain

import "time"

type Document struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Text        string    `json:"text"`
	ContentHash string    `json:"content_hash"`
	StatusCode  int       `json:"status_code"`
	Depth       int       `json:"depth"`
	FetchedAt   time.Time `json:"fetched_at"`
}
