package gomments

import (
	"context"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/aquilax/tripcode"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Service {
	return &Service{
		db: db,
	}
}

func getAuthorNameFallback(s string) string {
	if s == "" {
		return "Anonymous"
	}
	return s
}

func getReplySignatureFallback(s string) string {
	if s == "" {
		return ""
	}

	return tripcode.Tripcode(s) // only first 8 bytes of s are used
}

var reNewlines1 = regexp.MustCompile("\n{1}\n*")
var reNewlines2 = regexp.MustCompile(`\n{2}\n*`)

func stripConsecutiveWhitespace(s string) string {
	linesTrimmed := []string{}
	for line := range strings.Lines(s) {
		linesTrimmed = append(linesTrimmed, strings.TrimRightFunc(line, unicode.IsSpace))
	}

	return reNewlines2.ReplaceAllString(strings.Join(linesTrimmed, "\n"), strings.Repeat("\n", 2))
}

type GetRepliesRequest struct {
	Article string
}

type GetRepliesResponse struct {
	Replies Replies `json:"replies"`
}

func (s *Service) GetReplies(ctx context.Context, req GetRepliesRequest) (*GetRepliesResponse, ServiceError) {
	resp := &GetRepliesResponse{}

	replies, err := getRepliesForArticle(ctx, s.db, req.Article)
	if err != nil {
		return nil, Errorf(http.StatusInternalServerError, "getting replies: %w", err)
	}

	resp.Replies = replies
	return resp, nil
}

type SubmitReplyRequest struct {
	IdempotencyKey  string `json:"idempotency_key"`
	SignatureSecret string `json:"signature_secret"`
	Article         string
	Body            string `json:"body"`
	AuthorName      string `json:"author_name"`
}

type SubmitReplyResponse struct {
	Reply Reply `json:"reply"`
}

func (s *Service) SubmitReply(ctx context.Context, req SubmitReplyRequest) (*SubmitReplyResponse, ServiceError) {
	replyAuthorName := reNewlines1.ReplaceAllString(strings.TrimSpace(req.AuthorName), " ")
	replyBody := stripConsecutiveWhitespace(req.Body)
	replyArticle := strings.TrimSpace(req.Article)

	if replyArticle == "" {
		return nil, Errorf(http.StatusBadRequest, "requires reply article")
	}

	if replyBody == "" {
		return nil, Errorf(http.StatusBadRequest, "requires reply body")
	}

	if len(replyBody) > 500 {
		return nil, Errorf(http.StatusBadRequest, "reply body max length 500 characters reached")
	}

	if len(replyAuthorName) > 24 {
		return nil, Errorf(http.StatusBadRequest, "reply author name max length 24 characters reached")
	}

	if _, err := uuid.Parse(req.IdempotencyKey); err != nil {
		return nil, Errorf(http.StatusBadRequest, "parsing idempotency key: %w", err)
	}

	params := insertReplyParams{
		Article:        replyArticle,
		Body:           html.EscapeString(replyBody),
		Signature:      getReplySignatureFallback(req.SignatureSecret),
		IdempotencyKey: req.IdempotencyKey,
		AuthorName:     html.EscapeString(getAuthorNameFallback(replyAuthorName)),
		CreatedAt:      time.Now(),
	}
	replyID := 0
	if id, err := insertReply(
		ctx,
		s.db,
		params,
	); err != nil {
		return nil, Errorf(http.StatusInternalServerError, "inserting reply: %w", err)
	} else {
		replyID = id
	}

	return &SubmitReplyResponse{
		Reply: Reply{
			ID:             replyID,
			IdempotencyKey: params.IdempotencyKey,
			Signature:      params.Signature,
			Article:        params.Article,
			Body:           params.Body,
			Deleted:        params.Deleted,
			CreatedAt:      params.CreatedAt,
			AuthorName:     params.AuthorName,
		},
	}, nil
}

type GetStatsByArticlesRequest struct {
	Articles []string
}

type ArticleStats struct {
	Count       int       `json:"count"`
	LastReplyAt time.Time `json:"last_reply_at"`
}

type GetStatsByArticlesResponse struct {
	Stats map[string]ArticleStats `json:"stats"`
}

func (s *Service) GetStatsByArticles(ctx context.Context, req GetStatsByArticlesRequest) (*GetStatsByArticlesResponse, ServiceError) {
	aggs, err := getStatsForArticles(ctx, s.db, req.Articles)
	if err != nil {
		return nil, Errorf(http.StatusInternalServerError, "getting aggs: %w", err)
	}

	resp := &GetStatsByArticlesResponse{
		Stats: map[string]ArticleStats{},
	}

	for _, article := range req.Articles {
		resp.Stats[article] = ArticleStats{}
	}

	for _, agg := range aggs {
		resp.Stats[agg.Article] = ArticleStats{
			Count:       agg.Count,
			LastReplyAt: agg.LastReplyAt,
		}
	}

	return resp, nil
}
