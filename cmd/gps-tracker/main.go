package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Mock GPS tracking provider for Dosu Logi.
// Exposes REST API (X-Api-Key) and pushes webhooks with HMAC-SHA256 (X-Hmac-Signature).

type Vehicle struct {
	TrackingCode string    `json:"tracking_code"`
	Plate        string    `json:"plate"`
	Status       string    `json:"status"`
	Origin       string    `json:"origin"`
	Destination  string    `json:"destination"`
	Lat          float64   `json:"lat"`
	Lng          float64   `json:"lng"`
	Progress     float64   `json:"progress"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ShipmentEvent struct {
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	EventTime   time.Time `json:"event_time"`
}

type ShipmentResponse struct {
	TrackingCode string          `json:"tracking_code"`
	Status       string          `json:"status"`
	Origin       string          `json:"origin"`
	Destination  string          `json:"destination"`
	Lat          float64         `json:"lat"`
	Lng          float64         `json:"lng"`
	Events       []ShipmentEvent `json:"events"`
}

type WebhookPayload struct {
	TrackingCode string     `json:"tracking_code"`
	Status       string     `json:"status"`
	Location     string     `json:"location"`
	Lat          float64    `json:"lat"`
	Lng          float64    `json:"lng"`
	EventTime    *time.Time `json:"event_time"`
	Description  string     `json:"description"`
}

type Store struct {
	mu       sync.RWMutex
	vehicles map[string]*Vehicle
}

func main() {
	apiKey := env("GPS_API_KEY", "dev-gps-api-key")
	webhookURL := env("GPS_WEBHOOK_URL", "http://127.0.0.1:8089/api/v1/webhooks/tracking")
	webhookSecret := env("GPS_WEBHOOK_SECRET", "dev-tracking-secret")
	port := env("GPS_PORT", "8091")
	intervalSec := envInt("GPS_PUSH_INTERVAL_SEC", 30)

	store := &Store{vehicles: seedVehicles()}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/shipments", auth(apiKey, func(w http.ResponseWriter, _ *http.Request) {
		store.mu.RLock()
		defer store.mu.RUnlock()
		list := make([]ShipmentResponse, 0, len(store.vehicles))
		for _, v := range store.vehicles {
			list = append(list, store.toResponse(v))
		}
		writeJSON(w, http.StatusOK, list)
	}))
	mux.HandleFunc("/v1/shipments/", auth(apiKey, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Path[len("/v1/shipments/"):]
		if code == "" {
			http.Error(w, "tracking_code required", http.StatusBadRequest)
			return
		}
		store.mu.RLock()
		v, ok := store.vehicles[code]
		store.mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, store.toResponse(v))
	}))

	srv := &http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		log.Printf("gps-tracker listening on :%s (webhook=%s)", port, webhookURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go store.simulate(ctx, time.Duration(intervalSec)*time.Second, webhookURL, webhookSecret)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}

func seedVehicles() map[string]*Vehicle {
	now := time.Now()
	return map[string]*Vehicle{
		"VD-DEMO-001": {
			TrackingCode: "VD-DEMO-001", Plate: "51C-123.45", Status: "in_transit",
			Origin: "TP. Hồ Chí Minh", Destination: "Đà Nẵng",
			Lat: 13.85, Lng: 109.05, Progress: 0.35, UpdatedAt: now,
		},
		"VD-DEMO-002": {
			TrackingCode: "VD-DEMO-002", Plate: "29H-678.90", Status: "in_transit",
			Origin: "Hà Nội", Destination: "Hải Phòng",
			Lat: 20.95, Lng: 105.80, Progress: 0.55, UpdatedAt: now,
		},
		"VD-DEMO-003": {
			TrackingCode: "VD-DEMO-003", Plate: "43A-111.22", Status: "picked_up",
			Origin: "Cần Thơ", Destination: "An Giang",
			Lat: 10.10, Lng: 105.70, Progress: 0.10, UpdatedAt: now,
		},
	}
}

func (s *Store) toResponse(v *Vehicle) ShipmentResponse {
	now := time.Now()
	return ShipmentResponse{
		TrackingCode: v.TrackingCode,
		Status:       v.Status,
		Origin:       v.Origin,
		Destination:  v.Destination,
		Lat:          v.Lat,
		Lng:          v.Lng,
		Events: []ShipmentEvent{{
			Status:      v.Status,
			Description: fmt.Sprintf("Xe %s đang trên tuyến %s → %s", v.Plate, v.Origin, v.Destination),
			Location:    fmt.Sprintf("%.4f, %.4f", v.Lat, v.Lng),
			Lat:         v.Lat,
			Lng:         v.Lng,
			EventTime:   now,
		}},
	}
}

func (s *Store) simulate(ctx context.Context, every time.Duration, webhookURL, secret string) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			for _, v := range s.vehicles {
				v.Progress += 0.03 + rand.Float64()*0.04
				if v.Progress >= 1 {
					v.Progress = 1
					v.Status = "delivered"
				} else if v.Progress > 0.75 {
					v.Status = "out_for_delivery"
				} else if v.Progress > 0.2 {
					v.Status = "in_transit"
				}
				v.Lat, v.Lng = interpolateRoute(v.Progress, v.Origin, v.Destination)
				v.UpdatedAt = time.Now()
				go pushWebhook(webhookURL, secret, WebhookPayload{
					TrackingCode: v.TrackingCode,
					Status:       v.Status,
					Location:     fmt.Sprintf("%s (%.4f, %.4f)", v.Plate, v.Lat, v.Lng),
					Lat:          v.Lat,
					Lng:          v.Lng,
					EventTime:    ptrTime(v.UpdatedAt),
					Description:  fmt.Sprintf("Cập nhật GPS xe %s", v.Plate),
				})
			}
			s.mu.Unlock()
		}
	}
}

func interpolateRoute(p float64, origin, dest string) (lat, lng float64) {
	routes := map[string][2][2]float64{
		key("TP. Hồ Chí Minh", "Đà Nẵng"):   {{10.7769, 106.7009}, {16.0544, 108.2022}},
		key("Hà Nội", "Hải Phòng"):          {{21.0285, 105.8542}, {20.8449, 106.6881}},
		key("Cần Thơ", "An Giang"):          {{10.0452, 105.7469}, {10.5216, 105.1259}},
	}
	pair, ok := routes[key(origin, dest)]
	if !ok {
		return 10.8 + p*5, 106.6 + p*1.5
	}
	lat = pair[0][0] + (pair[1][0]-pair[0][0])*p
	lng = pair[0][1] + (pair[1][1]-pair[0][1])*p
	lat += (rand.Float64() - 0.5) * 0.02
	lng += (rand.Float64() - 0.5) * 0.02
	return lat, lng
}

func key(a, b string) string { return a + "|" + b }

func pushWebhook(url, secret string, payload WebhookPayload) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hmac-Signature", signHMAC(body, secret))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("webhook push failed %s: %v", payload.TrackingCode, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("webhook rejected %s: status %d", payload.TrackingCode, resp.StatusCode)
	}
}

func signHMAC(body []byte, secret string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func auth(apiKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func envInt(k string, fallback int) int {
	var n int
	if _, err := fmt.Sscanf(env(k, ""), "%d", &n); err == nil && n > 0 {
		return n
	}
	return fallback
}

func ptrTime(t time.Time) *time.Time { return &t }
