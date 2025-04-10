package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/sonroyaalmerol/sogo-tool-plus/internal/sogo"
)

type WebHandler struct {
	service *sogo.SogoService
}

func NewWebHandler(service *sogo.SogoService) *WebHandler {
	return &WebHandler{service: service}
}

func (h *WebHandler) HandleCalSubscribeUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[2] != "user" {
		http.Error(
			w,
			"Invalid URL path. Use /calendars/subscribe/user/{uid}",
			http.StatusBadRequest,
		)
		return
	}
	uid := parts[3]
	if uid == "" {
		http.Error(w, "User UID cannot be empty", http.StatusBadRequest)
		return
	}

	log.Printf("Received web request to subscribe user: %s", uid)
	err := h.service.CalSubscribeUser(uid)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Error processing web request for user %s: %v", uid, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf(
			"Subscription process initiated for user %s",
			uid,
		),
	})
}

func (h *WebHandler) HandleCalSubscribeAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Received web request to subscribe all users")
	err := h.service.CalSubscribeAll()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Error processing web request for subscribe-all: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subscription process initiated for all users",
	})
}
