package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransportSendPostsBearerAuthAndJSONBody(t *testing.T) {
	var got struct {
		method, auth, contentType, encoding string
		body                                batch
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.auth = r.Header.Get("Authorization")
		got.contentType = r.Header.Get("Content-Type")
		got.encoding = r.Header.Get("Content-Encoding")
		body := r.Body
		if got.encoding == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Fatalf("gzip reader: %v", err)
			}
			defer gz.Close()
			body = io.NopCloser(gz)
		}
		raw, _ := io.ReadAll(body)
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
	if got.encoding != "gzip" {
		t.Errorf("content-encoding=%q", got.encoding)
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

func TestTransportSendBodyIsGzipped(t *testing.T) {
	var encoding string
	var rawBody bytes.Buffer
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding = r.Header.Get("Content-Encoding")
		_, _ = io.Copy(&rawBody, r.Body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	tr := &transport{url: srv.URL, key: "k1", client: srv.Client()}
	if err := tr.send(batch{Host: "h", Metrics: []metric{{Name: "x", Value: 1}}}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if encoding != "gzip" {
		t.Errorf("encoding=%q want gzip", encoding)
	}
	gz, err := gzip.NewReader(&rawBody)
	if err != nil {
		t.Fatalf("not gzipped: %v", err)
	}
	defer gz.Close()
	decoded, _ := io.ReadAll(gz)
	if len(decoded) == 0 {
		t.Error("decoded body empty")
	}
}
