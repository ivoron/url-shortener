package handler

import (
	"encoding/json"
	"net/http"
	"net/url"

	"url-shortener/internal/service"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HTTPHandler struct {
	service service.ShortenerService
}

func NewHTTPHandler(service service.ShortenerService) *HTTPHandler {
	return &HTTPHandler{
		service: service,
	}
}

func (h *HTTPHandler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/shorten", h.ShortenURL)
	r.Get("/{shortURL}", h.ResolveURL)

	return r
}

type shortenRequest struct {
	OriginalUrl string `json:"url"`
}

type shortenResponse struct {
	ShortUrl string `json:"short_url"`
	Error    string `json:"error,omitempty"`
}

func (h *HTTPHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, shortenResponse{Error: "invalid request"})
		return
	}

	if _, err := url.ParseRequestURI(req.OriginalUrl); err != nil {
		respondJSON(w, http.StatusBadRequest, shortenResponse{Error: "Invalid URL"})
		return
	}

	urlEntity, err := h.service.Shorten(r.Context(), req.OriginalUrl)

	if err != nil {
		respondJSON(w, http.StatusInternalServerError, shortenResponse{Error: "Internal server error"})
		return
	}

	fullShortURL := "http://localhost:8080/" + urlEntity.ShortUrl

	respondJSON(w, http.StatusCreated, shortenResponse{ShortUrl: fullShortURL})
}

func (h *HTTPHandler) ResolveURL(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "shortURL")

	urlEntity, err := h.service.Resolve(r.Context(), code)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, urlEntity.OriginalUrl, http.StatusMovedPermanently)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	body, err := json.Marshal(payload)
	
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}