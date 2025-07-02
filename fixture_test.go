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
		gomments.New(dbx),
	}
}

type insertReplyParams struct {
	IdempotencyKey string `db:"reply_idempotency_key"`
	Signature      string `db:"reply_signature"`

	Article   string    `db:"reply_article"`
	Body      string    `db:"reply_body"`
	Deleted   bool      `db:"reply_deleted"`
	CreatedAt time.Time `db:"reply_created_at"`

	AuthorName string `db:"reply_author_name"`
}

func insertReply(ctx context.Context, db *sqlx.DB, params insertReplyParams) (int, error) {
	query := `
       INSERT INTO reply (
				   reply_idempotency_key,
				   reply_signature,
				   reply_article,
				   reply_body,
				   reply_deleted,
				   reply_created_at,
				   reply_author_name
       ) VALUES (
           :reply_idempotency_key,
           :reply_signature,
           :reply_article,
           :reply_body,
           :reply_deleted,
           :reply_created_at,
           :reply_author_name
       ) ON CONFLICT (reply_idempotency_key) DO UPDATE SET
				   reply_idempotency_key = excluded.reply_idempotency_key
			 RETURNING reply_id`

	q, args, err := db.BindNamed(query, params)
	if err != nil {
		return 0, fmt.Errorf("binding for insertReply: %w", err)
	}

	row := struct {
		ID int `db:"reply_id"`
	}{}

	if err := db.GetContext(ctx, &row, q, args...); err != nil {
		return 0, fmt.Errorf("selecting and inserting for insertReply: %w", err)
	}

	if row.ID == 0 {
		return 0, fmt.Errorf("unexpected id after insert")
	}

	return row.ID, nil
}
