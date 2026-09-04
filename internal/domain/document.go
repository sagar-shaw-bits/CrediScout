package domain

type Document struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Title string `json:"title"`
	Text  string `json:"text"`
}
