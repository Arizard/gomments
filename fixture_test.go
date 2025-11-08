package gomments_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/arizard/gomments"
	"github.com/arizard/gomments/internal"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	*require.Assertions

	db      *sqlx.DB
	service *gomments.Service
}

func newFixture(t *testing.T) fixture {
	ctx := context.Background()
	t.Parallel()

	dbPath := fmt.Sprintf("./data/gomments_test_%s.db", uuid.NewString())
	dbx, err := internal.InitSQLiteDatabase(dbPath)
	if err != nil {
		t.Error(err)
	}

	t.Cleanup(func() {
		os.Remove(dbPath)
	})

	return fixture{
		require.New(t),
		dbx,
		gomments.New(ctx, dbx),
	}
}

type insertReplyParams struct {
	IdempotencyKey string `db:"idempotency_key"`
	Signature      string `db:"signature"`

	Article   string    `db:"article"`
	Body      string    `db:"body"`
	Deleted   bool      `db:"deleted"`
	CreatedAt time.Time `db:"created_at"`

	AuthorName string `db:"author_name"`
}

func insertReply(ctx context.Context, db *sqlx.DB, params insertReplyParams) (int, error) {
	query := `
       INSERT INTO reply (
				   idempotency_key,
				   signature,
				   article,
				   body,
				   deleted,
				   created_at,
				   author_name
       ) VALUES (
           :idempotency_key,
           :signature,
           :article,
           :body,
           :deleted,
           :created_at,
           :author_name
       ) ON CONFLICT (idempotency_key) DO UPDATE SET
				   idempotency_key = excluded.idempotency_key
			 RETURNING id`

	q, args, err := db.BindNamed(query, params)
	if err != nil {
		return 0, fmt.Errorf("binding for insertReply: %w", err)
	}

	row := struct {
		ID int `db:"id"`
	}{}

	if err := db.GetContext(ctx, &row, q, args...); err != nil {
		return 0, fmt.Errorf("selecting and inserting for insertReply: %w", err)
	}

	if row.ID == 0 {
		return 0, fmt.Errorf("unexpected id after insert")
	}

	return row.ID, nil
}
