package gomments

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Reply struct {
	ID             int    `db:"id" json:"id"`
	IdempotencyKey string `db:"idempotency_key" json:"idempotency_key"`
	Signature      string `db:"signature" json:"signature"`

	Article   string    `db:"article" json:"article"`
	Body      string    `db:"body" json:"body"`
	Deleted   bool      `db:"deleted" json:"deleted"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	AuthorName string `db:"author_name" json:"author_name"`
}

type Replies []Reply

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

func getRepliesForArticle(ctx context.Context, db *sqlx.DB, article string) (Replies, error) {
	result := Replies{}

	err := db.SelectContext(
		ctx,
		&result,
		`
		SELECT
			 id,
			 idempotency_key,
			 signature,
			 article,
			 body,
			 deleted,
			 created_at,
			 author_name
		FROM reply
		WHERE article = ? AND deleted == false
		ORDER BY created_at DESC
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
		Article     string `db:"article"`
		Count       int    `db:"count"`
		LastReplyAt string `db:"last_at"`
	}{}

	query := `
		SELECT
			article,
			COUNT(id) AS count,
			DATETIME(MAX(created_at)) AS last_at
		FROM reply
		WHERE article IN (?) AND deleted = false
		GROUP BY article
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
