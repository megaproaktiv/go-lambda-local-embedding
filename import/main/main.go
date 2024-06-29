package main

import (
	"context"
	"fmt"
	he "hugoembedding"
	"hugoembedding/localstore"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

var Connection *pgx.Conn
var ctx context.Context
var Log *slog.Logger

func init() {
	// Logging
	Log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(Log)

	// Postgres
	ctx = context.Background()

}

const Version = 1

func main() {

	directoryPath := "./testdata"

	db, err := localstore.Init()
	ctx := context.Background()

	if err != nil {
		fmt.Println("Error creating collection:", err)
	}
	err = filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		he.Logger.Debug("Processing",
			"path", path,
			"name", info.Name())

		fileName := info.Name()
		if fileName == "index.md" {
			localstore.ProcessIndex(path, 1, db, ctx)
		}
		localstore.Store(db)
		return nil
	})
	if err != nil {
		fmt.Println("Error walking directory:", err)
	}

}
