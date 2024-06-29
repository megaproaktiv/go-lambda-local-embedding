package localstore_test

import (
	"context"
	"hugoembedding/localstore"
	"testing"

	be "github.com/megaproaktiv/bedrockembedding/titan"
	"github.com/philippgille/chromem-go"
	"gotest.tools/v3/assert"
)

// "os"
func TestSimpleQuery(t *testing.T) {
	// Setup
	db := chromem.NewDB()
	path := "testdata/db.gob"
	t.Log("Using local test db ", path)
	db, err := localstore.Load(path)
	assert.NilError(t, err)

	// Test
	ctx := context.Background()
	var myEmbeddingFunc chromem.EmbeddingFunc
	myEmbeddingFunc = MyEmbeddingFunc

	c := db.GetCollection("knowledge-base", myEmbeddingFunc)
	t.Logf("Collection initialized, count documents: %v\n", c.Count())
	questions := [...]string{
		"Container on ECS ec2 does not start?",
		"How do i generate a presigned URL?",
		"What duration has a presigned URL?",
		"What tipps for tuning ZED editor do you have?",
	}
	for _, q := range questions {

		t.Logf("Question: %v\n", q)
		res, err := c.Query(ctx, q, 10, nil, nil)
		if err != nil {
			panic(err)
		}
		t.Logf("Output has size %v\n", len(res))
		verbose := true
		for _, r := range res {
			if verbose {
				t.Logf("ID: %v - Similarity: %2.2f / Title: %v\n ===\n %v\n ====\n", r.ID, r.Similarity, r.Metadata["title"], r.Content)
			} else {
				t.Logf("ID: %v - Similarity: %2.2f / Title: %v\n", r.ID, r.Similarity, r.Metadata["title"])

			}
		}
	}

}

func MyEmbeddingFunc(ctx context.Context, text string) ([]float32, error) {

	return be.FetchEmbedding(text)
}
