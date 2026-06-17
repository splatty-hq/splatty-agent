package main

import (
	"testing"
)

func TestParseLoadavg(t *testing.T) {
	one, five, fifteen, err := parseLoadavg("0.42 1.23 2.5 1/123 4567\n")
	if err != nil {
		t.Fatal(err)
	}
	if one != 0.42 || five != 1.23 || fifteen != 2.5 {
		t.Errorf("got %v %v %v", one, five, fifteen)
	}
}

func TestParseLoadavgErrorsOnShort(t *testing.T) {
	if _, _, _, err := parseLoadavg("0.42 1.23"); err == nil {
		t.Fatal("expected error")
	}
}
