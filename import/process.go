package hugoembedding

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	be "github.com/megaproaktiv/bedrockembedding/titan"
	"github.com/pgvector/pgvector-go"
	"gopkg.in/yaml.v2"
)

type Metadata struct {
	Title string
	Autor string `yaml:"author"`
	Tags  []string
	Date  string
}

// Call process and import into embedding
func ProcessIndex(path string, baseRef string, conn *pgx.Conn, ctx context.Context) error {
	Logger.Info("Processing Index", "path", path)
	// Get chunks from file
	markdownFileContent, err := os.ReadFile(path)
	chunks, err := Parse(markdownFileContent)
	if err != nil {
		Logger.Error("Error parsing markdown file", "error", err)
		return err
	}

	// Compress Chunks
	chunks, err = CompressChunks(chunks, 300)
	if err != nil {
		Logger.Error("Error compressing chunks", "error", err)
		return err
	}
	// Get Metadata
	meta, err := ExtractMetadata(path)
	title := meta.Title
	link := baseRef + Path2Link(path, 1, meta.Date)
	// Put chunks into database
	for i, chunk := range *chunks {
		content := chunk.Chunk
		singleEmbedding, err := be.FetchEmbedding(*content)
		if err != nil {
			panic(err)
		}
		context := content
		c := *chunks
		if i > 0 && i < len(c)-1 {
			a := *(c[i-1].Chunk)
			b := *(c[i].Chunk)
			c := *(c[i+1].Chunk)
			cs := (a + b + c)
			context = &cs
		}
		// 		_, err = conn.Exec(ctx, "CREATE TABLE pow (id bigserial PRIMARY KEY, content text, context text, link text, title text, embedding vector(1536))")

		sql := "INSERT INTO pow (content, context,title, link, embedding) VALUES ($1, $2, $3, $4, $5)"
		Logger.Debug("SQL", "sql", sql)
		_, err = conn.Exec(ctx, sql,
			content,
			context,
			title,
			link,
			pgvector.NewVector(singleEmbedding))
		if err != nil {
			panic(err)
		}
		fmt.Printf("Cluster %v: {%v}\n / [%v]\n", i+1, content, context)
	}
	return nil
}

// Todo
// Convert
// /Users/gglawe/Documents/projects/community/2024/pearls/content/post/2024/pyrightconfig-zed/index.md
// to
// 2024/pyrightconfig-zed/
func Path2Link(path string, conversionMethod int, metadate string) string {
	// find `content/post/` and give right part
	Logger.Debug("Path2Links", "path", path, "Method", conversionMethod)
	var parts []string
	var link string
	if conversionMethod == 1 {
		parts = strings.Split(path, "content/")
		Logger.Debug("Parts", "parts", parts[1])
		link = parts[1]
	}
	if conversionMethod == 2 {
		parts = strings.Split(path, "post/")
		innerParts := strings.Split(parts[1], "/")
		fileName := innerParts[len(innerParts)-1]
		// Stip suffix ".md" from filename
		fileName = strings.Replace(fileName, ".md", "", 1)
		fileName = fileName + ".html"
		// Parsing date from metadate, getting month in MM format
		month, err := TryParseDateMonth(metadate)
		if err != nil {
			Logger.Error("Error parsing date", "error", err)
			link = "/"
		} else {
			link = innerParts[0] + "/" + *month + "/" + fileName
		}
	}

	// remove index.md
	link = strings.Replace(link, "index.md", "", 1)
	return link
}

func ExtractMetadata(filePath string) (*Metadata, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	meta := &Metadata{}

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(meta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func TryParseDateMonth(dateStr string) (*string, error) {
	// Define a slice of date formats to try
	formats := []string{
		"Wed, 02 Jan 2006 15:04:05 -0700", // dd-Mmm-yyyy
		"2006-01-02",
	}

	// Try each format until one succeeds or all fail
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			month := t.Format("01")
			return &month, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date: %s", dateStr)
}
