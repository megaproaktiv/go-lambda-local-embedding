package hugoembedding

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func CreateTable(conn *pgx.Conn, ctx context.Context, version int, dropTable *bool) {
	_, err := conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		panic(err)
	}

	if *dropTable {
		slog.Info("Drop table, Create table")
		_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS pow")
		if err != nil {
			panic(err)
		}

		_, err = conn.Exec(ctx, "CREATE TABLE pow (id bigserial PRIMARY KEY, content text, context text, link text, title text, embedding vector(1536))")

		if err != nil {
			panic(err)
		}

		_, err = conn.Exec(ctx, "CREATE INDEX embedding_pow_idx ON pow USING ivfflat(embedding)")
		if err != nil {
			panic(err)
		}
	}
}
