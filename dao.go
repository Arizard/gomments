package gomments

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Reply struct {
	ReplyID             int    `db:"reply_id" json:"reply_id"`
	ReplyIdempotencyKey string `db:"reply_idempotency_key" json:"reply_idempotency_key"`
	ReplySignature      string `db:"reply_signature" json:"reply_signature"`

	ReplyArticle   string    `db:"reply_article" json:"reply_article"`
	ReplyBody      string    `db:"reply_body" json:"reply_body"`
	ReplyDeleted   bool      `db:"reply_deleted" json:"reply_deleted"`
	ReplyCreatedAt time.Time `db:"reply_created_at" json:"reply_created_at"`

	ReplyAuthorName string `db:"reply_author_name" json:"reply_author_name"`
}

type Replies []Reply

type insertReplyParams struct {
	ReplyIdempotencyKey string `db:"reply_idempotency_key"`
	ReplySignature      string `db:"reply_signature"`

	ReplyArticle   string    `db:"reply_article"`
	ReplyBody      string    `db:"reply_body"`
	ReplyDeleted   bool      `db:"reply_deleted"`
	ReplyCreatedAt time.Time `db:"reply_created_at"`

	ReplyAuthorName string `db:"reply_author_name"`
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
		ReplyID int `db:"reply_id"`
	}{}

	if err := db.GetContext(ctx, &row, q, args...); err != nil {
		return 0, fmt.Errorf("selecting and inserting for insertReply: %w", err)
	}

	if row.ReplyID == 0 {
		return 0, fmt.Errorf("unexpected id after insert")
	}

	return row.ReplyID, nil
}

func getRepliesForArticle(ctx context.Context, db *sqlx.DB, article string) (Replies, error) {
	result := Replies{}

	err := db.SelectContext(
		ctx,
		&result,
		`
		SELECT
			 reply_id,
			 reply_idempotency_key,
			 reply_signature,
			 reply_article,
			 reply_body,
			 reply_deleted,
			 reply_created_at,
			 reply_author_name
		FROM reply
		WHERE reply_article = ? AND reply_deleted == false
		ORDER BY reply_created_at DESC
		`,
		article,
	)
	if err != nil {
		return result, err
	}

	return result, nil
}
