package handler

import (
	"encoding/json"
	"testing"
)

func TestScheduleBodyToInput(t *testing.T) {
	var b scheduleBody
	raw := `{"name":"ChatGPT","type":"expense","amount":14500,"accountId":3,"frequency":"monthly",
		"nextRunAt":"2026-10-08T09:00:00+08:00","endAt":"","amountPending":true,
		"foreignCurrency":"USD","foreignAmount":2000}`
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatal(err)
	}
	in := b.toInput()
	if in.Name != "ChatGPT" || in.Amount != 14500 || in.AccountID != 3 || !in.AmountPending ||
		in.ForeignCurrency != "USD" || in.ForeignAmount != 2000 {
		t.Fatalf("fields lost: %+v", in)
	}
	if in.NextRunAt.IsZero() {
		t.Fatal("nextRunAt not parsed")
	}
	if in.EndAt != nil {
		t.Fatal("empty endAt should be nil")
	}

	b.EndAt = "2027-01-01"
	if b.toInput().EndAt == nil {
		t.Fatal("endAt not parsed")
	}
}

func TestToPlanInputs(t *testing.T) {
	if toPlanInputs(nil) != nil {
		t.Fatal("nil plans must stay nil so existing plans are kept")
	}
	empty := []installmentPlanBody{}
	if got := toPlanInputs(&empty); got == nil || len(*got) != 0 {
		t.Fatal("empty slice must clear plans")
	}
	src := []installmentPlanBody{{Name: "手机", PrincipalFen: 600000, Periods: 12, FirstDueOn: "2026-11-05"}}
	got := toPlanInputs(&src)
	if len(*got) != 1 || (*got)[0].PrincipalFen != 600000 || (*got)[0].Periods != 12 {
		t.Fatalf("unexpected: %+v", *got)
	}
}
