package gomments

import (
	"context"
	"crypto/sha256"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type service struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *service {
	return &service{
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

	return fmt.Sprintf("%x", sha256.Sum256(fmt.Appendf(nil, "gomments-reply-secret-%s", s)))
}

var reNewlines1 = regexp.MustCompile("\n{1}\n*")
var reNewlines2 = regexp.MustCompile(`\n{2}\n*`)

func stripConsecutiveWhitespace(s string) string {
	linesTrimmed := []string{}
	for line := range strings.Lines(s) {
		linesTrimmed = append(linesTrimmed, strings.TrimSpace(line))
	}

	return reNewlines2.ReplaceAllString(strings.Join(linesTrimmed, "\n"), strings.Repeat("\n", n))
}

type GetRepliesRequest struct {
	Article string
}

type GetRepliesResponse struct {
	Replies Replies
}

func (s *service) GetReplies(ctx context.Context, req GetRepliesRequest) (*GetRepliesResponse, ServiceError) {
	resp := &GetRepliesResponse{}

	replies, err := getRepliesForArticle(ctx, s.db, req.Article)
	if err != nil {
		return nil, Errorf(http.StatusInternalServerError, "getting replies: %w", err)
	}

	resp.Replies = replies
	return resp, nil
}

type SubmitReplyRequest struct {
	ReplyIdempotencyKey  string `json:"reply_idempotency_key"`
	ReplySignatureSecret string `json:"reply_signature_secret"`
	ReplyArticle         string
	ReplyBody            string `json:"reply_body"`
	ReplyAuthorName      string `json:"reply_author_name"`
}

type SubmitReplyResponse struct {
	Reply Reply
}

func (s *service) SubmitReply(ctx context.Context, req SubmitReplyRequest) (*SubmitReplyResponse, ServiceError) {
	replyAuthorName := reNewlines1.ReplaceAllString(strings.TrimSpace(req.ReplyAuthorName), " ")
	replyBody := strings.TrimSpace(stripConsecutiveWhitespace(req.ReplyBody))
	replyArticle := strings.TrimSpace(req.ReplyArticle)

	if replyArticle == "" {
		return nil, Errorf(http.StatusBadRequest, "requires reply article")
	}

	if replyBody == "" {
		return nil, Errorf(http.StatusBadRequest, "requires reply body")
	}

	if len(replyBody) > 500 {
		return nil, Errorf(http.StatusBadRequest, "reply body max length 500 characters reached")
	}

	if len(replyAuthorName) > 40 {
		return nil, Errorf(http.StatusBadRequest, "reply author name max length 40 characters reached")
	}

	if req.ReplySignatureSecret != "" {
		if len(req.ReplySignatureSecret) < 10 {
			return nil, Errorf(http.StatusBadRequest, "requires reply secret to be >= 10 characters, or 0")
		}
		if len(req.ReplySignatureSecret) > 40 {
			return nil, Errorf(http.StatusBadRequest, "requires reply secret to be <= 40 characters, or 0")
		}
	}

	if _, err := uuid.Parse(req.ReplyIdempotencyKey); err != nil {
		return nil, Errorf(http.StatusBadRequest, "parsing idempotency key: %w", err)
	}

	params := insertReplyParams{
		ReplyArticle:        replyArticle,
		ReplyBody:           html.EscapeString(replyBody),
		ReplySignature:      getReplySignatureFallback(req.ReplySignatureSecret),
		ReplyIdempotencyKey: req.ReplyIdempotencyKey,
		ReplyAuthorName:     html.EscapeString(getAuthorNameFallback(replyAuthorName)),
		ReplyCreatedAt:      time.Now(),
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
			ReplyID:             replyID,
			ReplyIdempotencyKey: params.ReplyIdempotencyKey,
			ReplySignature:      params.ReplySignature,
			ReplyArticle:        params.ReplyArticle,
			ReplyBody:           params.ReplyBody,
			ReplyDeleted:        params.ReplyDeleted,
			ReplyCreatedAt:      params.ReplyCreatedAt,
			ReplyAuthorName:     params.ReplyAuthorName,
		},
	}, nil
}
