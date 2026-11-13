package main

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantFish   int
		wantPlants int
	}{
		{name: "automatic", wantFish: -1, wantPlants: -1},
		{name: "fish only", args: []string{"--fish", "30"}, wantFish: 30, wantPlants: -1},
		{name: "plants only", args: []string{"--plants=45"}, wantFish: -1, wantPlants: 45},
		{name: "empty", args: []string{"--fish=0", "--plants", "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts, err := parseOptions(tt.args, &bytes.Buffer{})
			if err != nil {
				t.Fatal(err)
			}
			for name, check := range map[string]struct {
				got  *int
				want int
			}{
				"fish":   {counts.fishCount, tt.wantFish},
				"plants": {counts.plantCount, tt.wantPlants},
			} {
				if check.want == -1 {
					if check.got != nil {
						t.Fatalf("%s should be automatic, got %d", name, *check.got)
					}
				} else if check.got == nil || *check.got != check.want {
					t.Fatalf("%s override does not match %d", name, check.want)
				}
			}
		})
	}
}

func TestParseOptionsErrorsAndHelp(t *testing.T) {
	for _, args := range [][]string{
		{"--fish=-1"}, {"--plants=-1"}, {"--fish=abc"}, {"--plants=1.5"},
		{"--fish"}, {"--unknown"}, {"unexpected"}, {"--plants=999999999999999999999999"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if _, err := parseOptions(args, &bytes.Buffer{}); err == nil {
				t.Fatal("invalid arguments accepted")
			}
		})
	}
	var output bytes.Buffer
	if _, err := parseOptions([]string{"--help"}, &output); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help returned %v", err)
	}
	for _, option := range []string{"-fish", "-plants", "automatic"} {
		if !strings.Contains(output.String(), option) {
			t.Fatalf("help missing %q: %s", option, output.String())
		}
	}
}
