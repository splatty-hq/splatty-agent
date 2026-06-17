package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransportSendPostsBearerAuthAndJSONBody(t *testing.T) {
	var got struct {
		method, auth, contentType string
		body                      batch
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.auth = r.Header.Get("Authorization")
		got.contentType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &got.body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	tr := &transport{url: srv.URL, key: "k1", client: srv.Client()}
	err := tr.send(batch{
		Host:      "h",
		Timestamp: time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC),
		Metrics:   []metric{{Name: "cpu.usage_percent", Value: 12.5}},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if got.method != "POST" {
		t.Errorf("method=%q", got.method)
	}
	if got.auth != "Bearer k1" {
		t.Errorf("auth=%q", got.auth)
	}
	if got.contentType != "application/json" {
		t.Errorf("content-type=%q", got.contentType)
	}
	if got.body.Host != "h" || len(got.body.Metrics) != 1 || got.body.Metrics[0].Name != "cpu.usage_percent" {
		t.Errorf("body=%+v", got.body)
	}
}

func TestTransportSendNon2xxIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	tr := &transport{url: srv.URL, key: "k1", client: srv.Client()}
	if err := tr.send(batch{Host: "h", Metrics: []metric{{Name: "x", Value: 1}}}); err == nil {
		t.Fatal("expected error on 403")
	}
}
