package db

import (
	"context"
	"log"
	"runtime"

	"github.com/jackc/pgx/v5"
)

func FinalizeTx(ctx context.Context, tx pgx.Tx, err *error) {
	fn := "unknown"
	if pc, _, _, ok := runtime.Caller(1); ok {
		if f := runtime.FuncForPC(pc); f != nil {
			fn = f.Name()
		}
	}

	if *err != nil {
		log.Printf("%s: rolling back transaction due to error: %v", fn, *err)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Printf("%s: rollback transaction failed: %v", fn, rbErr)
		}
		return
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		log.Printf("%s: commit transaction failed: %v", fn, commitErr)
	}
}
