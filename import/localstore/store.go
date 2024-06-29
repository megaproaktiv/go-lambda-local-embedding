package localstore

import (
	"hugoembedding"

	"github.com/philippgille/chromem-go"
)

// Store Database
func Store(db *chromem.DB) error {
	log := hugoembedding.Logger
	const path = "db-data/db.gob"
	log.Info("Storing Database", "path", path)
	return db.Export(path, false, "")
}
