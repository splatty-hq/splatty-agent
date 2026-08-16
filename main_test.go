package main

import "testing"

func TestVersionRequested(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"--version"}, true},
		{[]string{"-version"}, true},
		{[]string{"-v"}, true},
		{[]string{"--help"}, false},
		{[]string{"--verbose"}, false},
	}
	for _, c := range cases {
		if got := versionRequested(c.args); got != c.want {
			t.Errorf("versionRequested(%v)=%v, want %v", c.args, got, c.want)
		}
	}
}
