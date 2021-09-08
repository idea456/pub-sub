package monitoring

type Message struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Path    string `json:"path"`
}
