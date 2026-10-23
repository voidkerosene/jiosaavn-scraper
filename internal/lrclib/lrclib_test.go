package lrclib

import "testing"

func TestBestMatchPicksClosestDuration(t *testing.T) {
	results := []Lyrics{
		{ID: 1, DurationFloat: 100},
		{ID: 2, DurationFloat: 204},
		{ID: 3, DurationFloat: 500},
	}
	got := bestMatch(results, 203)
	if got == nil || got.ID != 2 {
		t.Fatalf("expected id 2 (closest to 203s), got %+v", got)
	}
	_ = got.Duration()
}

func TestBestMatchRejectsDifferentRecording(t *testing.T) {
	if got := bestMatch([]Lyrics{{ID: 1, DurationFloat: 300}}, 100); got != nil {
		t.Fatalf("expected nil for a 200s-off match, got %+v", got)
	}
}

func TestBestMatchEmptyAndUnknownDuration(t *testing.T) {
	if got := bestMatch(nil, 100); got != nil {
		t.Fatalf("expected nil for empty results, got %+v", got)
	}
	if got := bestMatch([]Lyrics{{ID: 7}}, 0); got == nil || got.ID != 7 {
		t.Fatalf("expected first result with unknown duration, got %+v", got)
	}
}
