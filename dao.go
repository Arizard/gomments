package gomments

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Reply struct {
	ID             int    `db:"reply_id" json:"reply_id"`
	IdempotencyKey string `db:"reply_idempotency_key" json:"reply_idempotency_key"`
	Signature      string `db:"reply_signature" json:"reply_signature"`

	Article   string    `db:"reply_article" json:"reply_article"`
	Body      string    `db:"reply_body" json:"reply_body"`
	Deleted   bool      `db:"reply_deleted" json:"reply_deleted"`
	CreatedAt time.Time `db:"reply_created_at" json:"reply_created_at"`

	AuthorName string `db:"reply_author_name" json:"reply_author_name"`
}

type Replies []Reply

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

type ReplyAggregation struct {
	Article     string
	Count       int
	LastReplyAt time.Time
}

type ReplyAggregations []ReplyAggregation

func getStatsForArticles(ctx context.Context, db *sqlx.DB, articles []string) (ReplyAggregations, error) {
	results := []struct {
		Article     string `db:"reply_article"`
		Count       int    `db:"reply_count"`
		LastReplyAt string `db:"last_reply_at"`
	}{}

	query := `
		SELECT
			reply_article,
			COUNT(reply_id) AS reply_count,
			DATETIME(MAX(reply_created_at)) AS last_reply_at
		FROM reply
		WHERE reply_article IN (?) AND reply_deleted = false
		GROUP BY reply_article
	`

	query, args, err := sqlx.In(query, articles)
	if err != nil {
		return nil, fmt.Errorf("interpolating IN: %w", err)
	}

	if err := db.SelectContext(
		ctx,
		&results,
		query,
		args...,
	); err != nil {
		return nil, fmt.Errorf("aggregating reply content: %w", err)
	}

	aggs := ReplyAggregations{}

	for _, result := range results {
		agg := ReplyAggregation{}
		agg.Article = result.Article
		agg.Count = result.Count
		agg.LastReplyAt, err = time.Parse(time.DateTime, result.LastReplyAt)
		if err != nil {
			return nil, fmt.Errorf("parsing time string into time.Time: %w", err)
		}

		aggs = append(aggs, agg)
	}

	return aggs, nil
}
