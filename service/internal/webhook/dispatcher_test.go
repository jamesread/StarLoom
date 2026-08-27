package webhook

import "testing"

func TestNormalizeEvent(t *testing.T) {
	e, err := NormalizeEvent("stars.awarded")
	if err != nil || e != "stars.awarded" {
		t.Fatalf("%v %v", e, err)
	}
	if _, err := NormalizeEvent("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeURL(t *testing.T) {
	if _, err := NormalizeURL("https://example.com/hook"); err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeURL("ftp://x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeEvents(t *testing.T) {
	got, err := NormalizeEvents([]string{"stars.awarded", " stars.awarded "})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "stars.awarded" {
		t.Fatalf("got=%v", got)
	}
}

func TestSignature(t *testing.T) {
	sig := Signature(`{"a":1}`, "secret")
	if len(sig) != 64 {
		t.Fatalf("sig=%s", sig)
	}
	for _, c := range sig {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Fatalf("non-hex sig=%s", sig)
		}
	}
}
