package service

import "testing"
import "time"

func TestIsDecemberThirdWeek(t *testing.T) {
	cases := []struct {
		day  int
		want bool
	}{
		{1, false}, {14, false}, {15, true}, {18, true}, {21, true}, {22, false}, {31, false},
	}
	for _, c := range cases {
		at := time.Date(2026, 12, c.day, 12, 0, 0, 0, time.Local)
		if got := IsDecemberThirdWeek(at); got != c.want {
			t.Fatalf("Dec %d: want %v got %v", c.day, c.want, got)
		}
	}
	at := time.Date(2026, 11, 18, 0, 0, 0, 0, time.Local)
	if IsDecemberThirdWeek(at) {
		t.Fatal("November should not match")
	}
}
