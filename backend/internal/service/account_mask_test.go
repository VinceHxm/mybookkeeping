package service

import "testing"

func TestMaskCardNo(t *testing.T) {
	if got := MaskCardNo("6222021234567890"); got != "············7890" {
		t.Fatalf("got %q", got)
	}
	if got := MaskCardNo("1234"); got != "1234" {
		t.Fatalf("short got %q", got)
	}
}

func TestMaskHolderName(t *testing.T) {
	if got := MaskHolderName("张三丰"); got != "张**" {
		t.Fatalf("got %q", got)
	}
}

func TestSanitizeSecretWrite(t *testing.T) {
	if _, ok := SanitizeSecretWrite("········7890"); ok {
		t.Fatal("masked card should not apply")
	}
	if _, ok := SanitizeSecretWrite("张**"); ok {
		t.Fatal("masked holder should not apply")
	}
	v, ok := SanitizeSecretWrite("6222021234567890")
	if !ok || v != "6222021234567890" {
		t.Fatalf("plain card %v %q", ok, v)
	}
	v, ok = SanitizeSecretWrite("")
	if !ok || v != "" {
		t.Fatalf("empty clear %v %q", ok, v)
	}
}
