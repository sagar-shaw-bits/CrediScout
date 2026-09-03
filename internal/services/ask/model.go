package ask

type AskRequest struct {
	Question string `json:"question" binding:"required"`
	TopK     int    `json:"top_k"`
}
type Source struct {
	URL   string  `json:"url"`
	Title string  `json:"title"`
	Score float32 `json:"score"`
}

type AskResponse struct {
	Answer  string   `json:"answer"`
	Sources []Source `json:"sources"`
}
