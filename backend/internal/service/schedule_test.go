package service

import (
	"testing"
	"time"
)

func TestNextRunMonthlyAnchoredToMonthEnd(t *testing.T) {
	loc := time.Local
	at := time.Date(2026, 1, 31, 9, 0, 0, 0, loc)
	want := []string{"2026-02-28", "2026-03-31", "2026-04-30", "2026-05-31"}
	for _, w := range want {
		at = nextRun(at, "monthly", 1, 31)
		if got := at.Format("2006-01-02"); got != w {
			t.Fatalf("got %s want %s", got, w)
		}
		if at.Hour() != 9 {
			t.Fatalf("time of day lost: %v", at)
		}
	}
}

func TestNextRunYearlyLeapDay(t *testing.T) {
	loc := time.Local
	at := time.Date(2028, 2, 29, 0, 0, 0, 0, loc)
	at = nextRun(at, "yearly", 1, 29)
	if got := at.Format("2006-01-02"); got != "2029-02-28" {
		t.Fatalf("got %s", got)
	}
	at = nextRun(at, "yearly", 3, 29)
	if got := at.Format("2006-01-02"); got != "2032-02-29" {
		t.Fatalf("got %s", got)
	}
}

func TestNextRunEveryNDays(t *testing.T) {
	loc := time.Local
	at := time.Date(2026, 1, 31, 0, 0, 0, 0, loc)
	at = nextRun(at, "every_n_days", 30, 0)
	if got := at.Format("2006-01-02"); got != "2026-03-02" {
		t.Fatalf("got %s", got)
	}
}

func TestNextRunMonthlyEveryNMonths(t *testing.T) {
	loc := time.Local
	at := time.Date(2026, 11, 15, 0, 0, 0, 0, loc)
	at = nextRun(at, "monthly", 3, 0)
	if got := at.Format("2006-01-02"); got != "2027-02-15" {
		t.Fatalf("got %s", got)
	}
}
