package gomments

import (
	"context"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/aquilax/tripcode"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	db       *sqlx.DB
	sessions sync.Map
}

func New(ctx context.Context, db *sqlx.DB) *Service {
	s := &Service{
		db: db,
	}

	return s
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

func (s *Service) GetReplies(ctx context.Context, req GetRepliesRequest) (*GetRepliesResponse, error) {
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

func (s *Service) SubmitReply(ctx context.Context, req SubmitReplyRequest) (*SubmitReplyResponse, error) {
	authorName := reNewlines1.ReplaceAllString(strings.TrimSpace(req.AuthorName), " ")
	body := stripConsecutiveWhitespace(req.Body)
	article := strings.TrimSpace(req.Article)

	if article == "" {
		return nil, Errorf(http.StatusBadRequest, "requires reply article")
	}

	if body == "" {
		return nil, Errorf(http.StatusBadRequest, "requires reply body")
	}

	if len(body) > 500 {
		return nil, Errorf(http.StatusBadRequest, "reply body max length 500 characters reached")
	}

	if len(authorName) > 24 {
		return nil, Errorf(http.StatusBadRequest, "reply author name max length 24 characters reached")
	}

	if _, err := uuid.Parse(req.IdempotencyKey); err != nil {
		return nil, Errorf(http.StatusBadRequest, "parsing idempotency key: %w", err)
	}

	params := insertReplyParams{
		Article:        article,
		Body:           body,
		Signature:      getReplySignatureFallback(req.SignatureSecret),
		IdempotencyKey: req.IdempotencyKey,
		AuthorName:     getAuthorNameFallback(authorName),
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

type GetReplyStatsByArticlesRequest struct {
	Articles []string
}

type ArticleReplyStats struct {
	Count       int       `json:"count"`
	LastReplyAt time.Time `json:"last_reply_at"`
}

type GetReplyStatsByArticlesResponse struct {
	Stats map[string]ArticleReplyStats `json:"stats"`
}

func (s *Service) GetReplyStatsByArticles(ctx context.Context, req GetReplyStatsByArticlesRequest) (*GetReplyStatsByArticlesResponse, error) {
	aggs, err := getReplyStatsByArticles(ctx, s.db, req.Articles)
	if err != nil {
		return nil, Errorf(http.StatusInternalServerError, "getting aggs: %w", err)
	}

	resp := &GetReplyStatsByArticlesResponse{
		Stats: map[string]ArticleReplyStats{},
	}

	for _, article := range req.Articles {
		resp.Stats[article] = ArticleReplyStats{}
	}

	for _, agg := range aggs {
		resp.Stats[agg.Article] = ArticleReplyStats{
			Count:       agg.Count,
			LastReplyAt: agg.LastReplyAt,
		}
	}

	return resp, nil
}

var ValidReactionKinds []string = []string{"THUMBS_UP"}

type CreateReactionRequest struct {
	Kind    string
	Article string
}

type CreateReactionResponse struct {
	DeletionKey string `json:"deletion_key"`
}

func (s *Service) CreateReaction(ctx context.Context, req CreateReactionRequest) (*CreateReactionResponse, error) {
	if !slices.Contains(ValidReactionKinds, req.Kind) {
		return nil, Errorf(400, "not a valid kind: %q", req.Kind)
	}
	deletionKey := uuid.New().String()
	err := insertReaction(ctx, s.db, req.Article, req.Kind, deletionKey)
	if err != nil {
		return nil, Errorf(500, "creating reaction: %w", err)
	}

	return &CreateReactionResponse{DeletionKey: deletionKey}, nil
}

type DeleteReactionRequest struct {
	DeletionKey string
}

type DeleteReactionResponse struct {
}

func (s *Service) DeleteReaction(ctx context.Context, req DeleteReactionRequest) (*DeleteReactionResponse, error) {
	err := deleteReactionByDeletionKey(ctx, s.db, req.DeletionKey)
	if err != nil {
		return nil, Errorf(500, "deleting reaction: %w", err)
	}

	return &DeleteReactionResponse{}, nil
}

type GetReactionStatsByArticlesRequest struct {
	Articles []string
}

type ArticleReactionStats map[string]int
type GetReactionStatsByArticlesResponse struct {
	Stats map[string]ArticleReactionStats `json:"stats"`
}

func (s *Service) GetReactionStatsByArticles(ctx context.Context, req GetReactionStatsByArticlesRequest) (*GetReactionStatsByArticlesResponse, error) {
	aggs, err := getReactionStatsByArticles(ctx, s.db, req.Articles)
	if err != nil {
		return nil, Errorf(500, "aggregating reactions: %w", err)
	}

	resp := &GetReactionStatsByArticlesResponse{
		Stats: map[string]ArticleReactionStats{},
	}

	for _, article := range req.Articles {
		resp.Stats[article] = ArticleReactionStats{}
		for _, kind := range ValidReactionKinds {
			resp.Stats[article][kind] = 0
		}
	}

	for _, agg := range aggs {
		resp.Stats[agg.Article][agg.Kind] = agg.Count
	}

	return resp, nil
}
