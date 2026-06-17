package main

import "time"

type metric struct {
	Name  string            `json:"name"`
	Value float64           `json:"value"`
	Tags  map[string]string `json:"tags,omitempty"`
}

type batch struct {
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
	Metrics   []metric  `json:"metrics"`
}

type collector interface {
	collect() ([]metric, error)
}
