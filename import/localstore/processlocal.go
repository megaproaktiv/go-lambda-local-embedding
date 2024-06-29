package localstore

import (
	"context"
	"fmt"
	he "hugoembedding"
	"os"
	"runtime"
	"strconv"

	be "github.com/megaproaktiv/bedrockembedding/titan"
	"github.com/philippgille/chromem-go"
)

var IDCount int

func init() {
	IDCount = 0
}

// Call process and import into embedding
func ProcessIndex(path string, conversionMethod int, db *chromem.DB, ctx context.Context) error {
	log := he.Logger

	log.Info("Processing Index", "path", path)

	collection := db.GetCollection("knowledge-base", nil)

	// Get chunks from file
	markdownFileContent, err := os.ReadFile(path)
	chunks, err := he.Parse(markdownFileContent)
	if err != nil {
		log.Error("Error parsing markdown file", "error", err)
		return err
	}

	// Compress Chunks
	chunks, err = he.CompressChunks(chunks, 300)
	if err != nil {
		log.Error("Error compressing chunks", "error", err)
		return err
	}
	// Get Metadata
	meta, err := he.ExtractMetadata(path)
	title := ""
	link := ""
	if meta != nil {
		title = meta.Title
	}
	if err != nil {
		log.Error("Metadata extraction problem:", "error", err, "file", path)
	}
	// Put chunks into database
	for i, chunk := range *chunks {
		content := chunk.Chunk

		context := content
		c := *chunks
		if i > 0 && i < len(c)-1 {
			a := *(c[i-1].Chunk)
			b := *(c[i].Chunk)
			c := *(c[i+1].Chunk)
			cs := (a + b + c)
			context = &cs
		}

		id := strconv.Itoa(IDCount)
		IDCount++

		metaData := map[string]string{
			"link":  link,
			"title": title,
		}
		log.Info("Adding document into chromem", "count", id, "content", *content, "link", link, "title", title)
		singleEmbedding, err := be.FetchEmbedding(*content)
		// ***** ID Must be unique *****
		// Other wise documents will be overwritten
		err = collection.AddDocuments(ctx, []chromem.Document{
			{
				ID:        id,
				Content:   *content,
				Embedding: singleEmbedding,
				Metadata:  metaData,
			},
		}, runtime.NumCPU())
		if err != nil {
			panic(err)
		}

		fmt.Printf("Cluster %v: {%v}\n / [%v]\n", i+1, content, context)
	}
	return nil
}
