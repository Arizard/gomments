package gomments_test

import (
	"context"
	"testing"
	"time"

	"github.com/arizard/gomments"
)

func TestService_GetReplies(t *testing.T) {
	now := time.Now()
	articles := []string{
		"cGhwLW91dHB1dC1idWZmZXJzCg==",
		"Y29tbWVudC1zeXN0ZW0tZGVtbwo=",
	}

	replies := gomments.Replies{
		{
			ID:             2,
			IdempotencyKey: "ab37a08d-043b-40cc-b6a8-e2553eedcae9",
			Article:        articles[0],
			Body:           "Comment 0",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Arie",
		},
		{
			ID:             3,
			IdempotencyKey: "63292c12-205e-4dbe-9d9a-3584a8403a2b",
			Article:        articles[0],
			Body:           "Comment 1",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Anonymous",
		},
		{
			ID:             4,
			IdempotencyKey: "ee72e57f-dbb3-400b-b42f-d6a23b950091",
			Article:        articles[0],
			Body:           "Comment 2",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Anonymous",
		},
		{
			ID:             5,
			IdempotencyKey: "492462c4-0a18-4888-b66f-de5affc47cb3",
			Article:        articles[0],
			Body:           "Comment 3",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Anonymous",
		},
		{
			ID:             6,
			IdempotencyKey: "d58b4a1d-920f-4da5-a3a3-28eb2140945c",
			Article:        articles[0],
			Body:           "Comment 4",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Arie",
		},
		{
			ID:             7,
			IdempotencyKey: "7c27493b-6d30-490a-9676-d1ca70eb26aa",
			Article:        articles[0],
			Body:           "Comment",
			Deleted:        true,
			CreatedAt:      now.Add(-time.Hour),
			AuthorName:     "Arie",
		},
		{
			ID:             8,
			IdempotencyKey: "da2c98e5-2487-4ef4-ac47-ccba217050e5",
			Article:        articles[1],
			Body:           "Comment",
			Deleted:        false,
			CreatedAt:      now.Add(-time.Hour),
			AuthorName:     "Arie",
		},
		{
			ID:             9,
			IdempotencyKey: "bc88c355-e9a7-4a7c-b7aa-4f92d338ec79",
			Article:        articles[1],
			Body:           "Comment",
			Deleted:        true,
			CreatedAt:      now.Add(-time.Hour),
			AuthorName:     "Arie",
		},
	}

	tests := []struct {
		name string // description of this test case
		req  gomments.GetRepliesRequest
		want *gomments.GetRepliesResponse
		err  error
	}{
		{
			name: "gets_replies",
			req: gomments.GetRepliesRequest{
				Article: articles[0],
			},
			want: &gomments.GetRepliesResponse{
				Replies: gomments.Replies{
					replies[0],
					replies[1],
					replies[2],
					replies[3],
					replies[4],
				},
			},
			err: nil,
		},
		{
			name: "gets_replies_2",
			req: gomments.GetRepliesRequest{
				Article: articles[1],
			},
			want: &gomments.GetRepliesResponse{
				Replies: gomments.Replies{
					replies[6],
				},
			},
			err: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			ctx := context.Background()
			f := newFixture(tt)
			s := f.service

			for _, reply := range replies {
				_, err := insertReply(ctx, f.db, insertReplyParams{
					IdempotencyKey: reply.IdempotencyKey,
					Signature:      reply.Signature,
					Article:        reply.Article,
					Body:           reply.Body,
					Deleted:        reply.Deleted,
					CreatedAt:      reply.CreatedAt,
					AuthorName:     reply.AuthorName,
				})
				f.NoError(err)
			}

			got, err := s.GetReplies(ctx, tc.req)
			if tc.err == nil {
				f.NoError(err)
			} else {
				f.Equal(tc.err, err)
			}
			f.EqualExportedValues(tc.want, got)
		})
	}
}

func TestService_GetStatsByArticles(t *testing.T) {
	now := time.Now()
	nowOneHourAgo := now.Add(-time.Hour)
	articles := []string{
		"cGhwLW91dHB1dC1idWZmZXJzCg==",
		"Y29tbWVudC1zeXN0ZW0tZGVtbwo=",
	}

	replies := gomments.Replies{
		{
			ID:             2,
			IdempotencyKey: "ab37a08d-043b-40cc-b6a8-e2553eedcae9",
			Article:        articles[0],
			Body:           "Comment 0",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Arie",
		},
		{
			ID:             3,
			IdempotencyKey: "63292c12-205e-4dbe-9d9a-3584a8403a2b",
			Article:        articles[0],
			Body:           "Comment 1",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Anonymous",
		},
		{
			ID:             4,
			IdempotencyKey: "ee72e57f-dbb3-400b-b42f-d6a23b950091",
			Article:        articles[0],
			Body:           "Comment 2",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Anonymous",
		},
		{
			ID:             5,
			IdempotencyKey: "492462c4-0a18-4888-b66f-de5affc47cb3",
			Article:        articles[0],
			Body:           "Comment 3",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Anonymous",
		},
		{
			ID:             6,
			IdempotencyKey: "d58b4a1d-920f-4da5-a3a3-28eb2140945c",
			Article:        articles[0],
			Body:           "Comment 4",
			Deleted:        false,
			CreatedAt:      now,
			AuthorName:     "Arie",
		},
		{
			ID:             7,
			IdempotencyKey: "7c27493b-6d30-490a-9676-d1ca70eb26aa",
			Article:        articles[0],
			Body:           "Comment",
			Deleted:        true,
			CreatedAt:      nowOneHourAgo,
			AuthorName:     "Arie",
		},
		{
			ID:             8,
			IdempotencyKey: "da2c98e5-2487-4ef4-ac47-ccba217050e5",
			Article:        articles[1],
			Body:           "Comment",
			Deleted:        false,
			CreatedAt:      nowOneHourAgo,
			AuthorName:     "Arie",
		},
		{
			ID:             9,
			IdempotencyKey: "bc88c355-e9a7-4a7c-b7aa-4f92d338ec79",
			Article:        articles[1],
			Body:           "Comment",
			Deleted:        true,
			CreatedAt:      nowOneHourAgo,
			AuthorName:     "Arie",
		},
	}

	tests := []struct {
		name string // description of this test case
		req  gomments.GetReplyStatsByArticlesRequest
		want *gomments.GetReplyStatsByArticlesResponse
		err  error
	}{
		{
			name: "gets_stats_article",
			req: gomments.GetReplyStatsByArticlesRequest{
				Articles: []string{articles[0], articles[1], "non-existing-article"},
			},
			want: &gomments.GetReplyStatsByArticlesResponse{
				Stats: map[string]gomments.ArticleReplyStats{
					articles[0]: {
						Count:       5,
						LastReplyAt: now,
					},
					articles[1]: {
						Count:       1,
						LastReplyAt: nowOneHourAgo,
					},
					"non-existing-article": {},
				},
			},
			err: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			ctx := context.Background()
			f := newFixture(tt)
			s := f.service

			for _, reply := range replies {
				_, err := insertReply(ctx, f.db, insertReplyParams{
					IdempotencyKey: reply.IdempotencyKey,
					Signature:      reply.Signature,
					Article:        reply.Article,
					Body:           reply.Body,
					Deleted:        reply.Deleted,
					CreatedAt:      reply.CreatedAt,
					AuthorName:     reply.AuthorName,
				})
				f.NoError(err)
			}

			got, err := s.GetReplyStatsByArticles(ctx, tc.req)
			if tc.err == nil {
				f.NoError(err)
			} else {
				f.Equal(tc.err, err)
			}
			f.EqualExportedValues(tc.want, got)
		})
	}
}

func TestService_CreateReaction(t *testing.T) {
	t.Run("creates reactions", func(tt *testing.T) {
		ctx := context.Background()
		f := newFixture(tt)
		s := f.service

		resp, err := s.CreateReaction(ctx, gomments.CreateReactionRequest{Article: "test-article", Kind: "THUMBS_UP"})
		f.NoError(err)
		f.NotEmpty(resp.DeletionKey)

		resp, err = s.CreateReaction(ctx, gomments.CreateReactionRequest{Article: "test-article", Kind: "THUMBS_UP"})
		f.NoError(err)
		f.NotEmpty(resp.DeletionKey)

		resp, err = s.CreateReaction(ctx, gomments.CreateReactionRequest{Article: "test-article", Kind: "THUMBS_UP"})
		f.NoError(err)
		f.NotEmpty(resp.DeletionKey)

		deleteResp, err := s.DeleteReaction(ctx, gomments.DeleteReactionRequest{DeletionKey: resp.DeletionKey})
		f.NotNil(deleteResp)
		f.NoError(err)

		_, err = s.CreateReaction(ctx, gomments.CreateReactionRequest{Article: "test-article", Kind: "THUMBS_DOWN"})
		f.NotNil(err)

		statsResp, err := s.GetReactionStatsByArticles(ctx, gomments.GetReactionStatsByArticlesRequest{Articles: []string{"test-article"}})
		f.NoError(err)
		f.Equal(2, statsResp.Stats["test-article"]["THUMBS_UP"])
	})
}
