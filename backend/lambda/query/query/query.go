package query

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"ragembeddings"
	"strconv"

	re "ragembeddings"
	"ragembeddings/bedrock"
	"ragembeddings/localstore"

	be "github.com/megaproaktiv/bedrockembedding/titan"

	"github.com/philippgille/chromem-go"
)

var db *chromem.DB

func init() {

	path := "./db.gob"
	var err error
	db, err = localstore.Load(path)
	if err != nil {
		panic(err)
	}
}
func Query(c context.Context, req re.QueryRequest) re.Response {

	log := re.Logger

	question := req.Question
	log.Info("Question received", "question", question)
	// log.Println("Category", req.Category)
	// log.Println("Version", req.Version)

	var myEmbeddingFunc chromem.EmbeddingFunc
	myEmbeddingFunc = MyEmbeddingFunc

	log.Info("Query collection start")
	collection := db.GetCollection("knowledge-base", myEmbeddingFunc)
	res, err := collection.Query(c, question, 5, nil, nil)
	if err != nil {
		panic(err)
	}

	// rows, err := collection.Query(c, "SELECT id, content,context,link, title  FROM documents ORDER BY embedding <=> $1 LIMIT 10", pgvector.NewVector(embedding))

	log.Info("Template creation start")
	promptTemplate, err := os.ReadFile("prompt.tmpl")
	var templateStr string
	templateStr = string(promptTemplate)
	content_separator := os.Getenv("CONTENT_SEPARATOR")
	if content_separator == "" {
		content_separator = "document"
	}
	preExcerpt := fmt.Sprintf("<%v>\n", content_separator)
	postExcerpt := fmt.Sprintf("</%v>\n", content_separator)

	documentExcerpts := ""
	Documents := make([]ragembeddings.RagDocument, 0)
	for _, r := range res {
		idString := r.ID
		id, err := strconv.Atoi(idString)
		if err != nil {
			id = 1
			log.Error("Wrong ID: ", "id", idString)
		}
		content := r.Content

		context := r.Metadata["link"]
		// link := r.Metadata["link"]
		// title := r.Metadata["title"]

		log.Debug("Found", "id", id, "content", content[:64])
		documentExcerpts += preExcerpt
		documentExcerpts += content + "\n"
		documentExcerpts += postExcerpt

		Documents = append(Documents, ragembeddings.RagDocument{
			Id:      id,
			Content: content,
			Context: context,
		})
	}
	tmpl, err := template.New("Prompt").Parse(templateStr)
	if err != nil {
		log.Error("Error parsing template:", err)
	}
	data := ragembeddings.TemplateData{
		Question: question,
		Document: documentExcerpts,
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		log.Error("Error executing template:", err)
	}

	// Extract the string from the buffer
	log.Info("Asking claude")
	prompt := buffer.String()
	answer := bedrock.Chat(prompt)
	response := ragembeddings.Response{
		Answer:    answer,
		Documents: Documents,
	}
	log.Info("Answer received claude")
	return response
}
func MyEmbeddingFunc(ctx context.Context, text string) ([]float32, error) {

	return be.FetchEmbedding(text)
}
