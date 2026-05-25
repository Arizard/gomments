package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"context"
	"net/http"

	"github.com/arizard/gomments"
	"github.com/arizard/gomments/internal"
)

func mustGetEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env variable: %s", k)
	}
	return v
}

func jsonResponse(w http.ResponseWriter, resp any) {
	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", b)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	settings := struct {
		port        string
		baseURL     string
		allowOrigin string
		// cors        cors.Config
	}{
		port:        mustGetEnv("PORT"),
		baseURL:     os.Getenv("BASE_URL"),
		allowOrigin: os.Getenv("ALLOW_ORIGIN"),
		// cors:        cors.DefaultConfig(),
	}

	log.Printf("base url is %q", settings.baseURL)

	dbx, err := internal.InitSQLiteDatabase("/home/appuser/data/gomments.db")
	if err != nil {
		log.Fatalf("getting migrated dbx: %s", err)
		return
	}

	svc := gomments.New(ctx, dbx)

	svcMux := http.NewServeMux()
	svcMux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	})
	svcMux.HandleFunc("GET /articles/{article}/replies", func(w http.ResponseWriter, r *http.Request) {
		article := r.PathValue("article")
		if article == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resp, err := svc.GetReplies(ctx, gomments.GetRepliesRequest{Article: article})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		jsonResponse(w, resp)
	})
	svcMux.HandleFunc("POST /articles/{article}/replies", func(w http.ResponseWriter, r *http.Request) {
		var req gomments.SubmitReplyRequest
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			fmt.Fprint(w, "could not read body")
			return
		}

		if err := json.Unmarshal(b, &req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			fmt.Fprint(w, "could not unmarshal json")
			return
		}

		req.Article = r.PathValue("article")

		resp, err := svc.SubmitReply(ctx, req)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		jsonResponse(w, resp)
	})
	svcMux.HandleFunc("GET /articles/replies/stats", func(w http.ResponseWriter, r *http.Request) {
		req := gomments.GetReplyStatsByArticlesRequest{
			Articles: r.URL.Query()["article"],
		}
		resp, err := svc.GetReplyStatsByArticles(ctx, req)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, resp)
	})
	svcMux.HandleFunc("GET /articles/reactions/stats", func(w http.ResponseWriter, r *http.Request) {
		req := gomments.GetReactionStatsByArticlesRequest{
			Articles: r.URL.Query()["article"],
		}
		resp, err := svc.GetReactionStatsByArticles(ctx, req)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, resp)
	})
	svcMux.HandleFunc("POST /articles/{article}/reactions/{kind}", func(w http.ResponseWriter, r *http.Request) {
		article := r.PathValue("article")
		kind := r.PathValue("kind")

		resp, err := svc.CreateReaction(ctx, gomments.CreateReactionRequest{
			Kind:    kind,
			Article: article,
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, resp)
	})
	svcMux.HandleFunc("DELETE /reactions", func(w http.ResponseWriter, r *http.Request) {
		resp, err := svc.DeleteReaction(ctx, gomments.DeleteReactionRequest{
			DeletionKey: r.URL.Query().Get("key"),
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, resp)
	})

	rootMux := http.NewServeMux()
	strippedMux := http.StripPrefix(settings.baseURL, svcMux)
	rootMux.HandleFunc(fmt.Sprintf("%s/", settings.baseURL), func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", strings.Join([]string{settings.allowOrigin}, ","))
		h.Set("Access-Control-Allow-Methods", "OPTIONS, HEAD, GET, POST, DELETE")
		h.Set("Access-Control-Allow-Headers", "*")
		switch r.Method {
		case "OPTIONS": // If this is behind nginx, should cors logic be implemented there instead? Makes it more annoying for dev testing though.
			w.WriteHeader(http.StatusNoContent)
		default:
			h.Set("Content-Type", "application/json")
			strippedMux.ServeHTTP(w, r)
		}
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%s", settings.port), rootMux); err != nil {
		log.Fatal(err)
	}
}
