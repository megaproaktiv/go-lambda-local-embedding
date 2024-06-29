package rag

type QueryRequest struct {
	Question string `json:"question"`
	License  string `json:"license,omitempty"`
}

type RagDocument struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	Context string `json:"context"`
}

type Response struct {
	Answer    string        `json:"answer"`
	Documents []RagDocument `json:"documents"`
}

type TemplateData struct {
	Question string
	Document string
}
