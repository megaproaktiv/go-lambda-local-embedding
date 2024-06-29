package localstore


import (
	"context"

	"ragembeddings"

	be "github.com/megaproaktiv/bedrockembedding/titan"
	"github.com/philippgille/chromem-go"
)

// Init local hugoembedding database
func Init() (*chromem.DB, error) {
	log := ragembeddings.Logger
	db := chromem.NewDB()

	var myEmbeddingFunc chromem.EmbeddingFunc
	myEmbeddingFunc = MyEmbeddingFunc
	_, err := db.CreateCollection("knowledge-base", nil, myEmbeddingFunc)
	if err != nil {
		log.Error("Error creating collection", "error", err)
		return nil, err
	}
	return db, nil
}

func Load(path string) (*chromem.DB, error) {
	log := ragembeddings.Logger
	db := chromem.NewDB()

	err := db.Import(path, "")

	if err != nil {
		log.Error("Error loading collection", "error", err)
		return nil, err
	}
	return db, nil
}

// type EmbeddingFunc func(ctx context.Context, text string) ([]float32, error)
func MyEmbeddingFunc(ctx context.Context, text string) ([]float32, error) {

	return be.FetchEmbedding(text)
}
