package auth

import "testing"

func TestLoginLimiterAllowsThenBlocks(t *testing.T) {
	l := NewLoginLimiter()
	for i := 0; i < 10; i++ {
		if !l.Allow("1.1.1.1", "t", "a@b.c") {
			t.Fatalf("attempt %d should allow", i+1)
		}
	}
	if l.Allow("1.1.1.1", "t", "a@b.c") {
		t.Fatal("11th should deny")
	}
	if !l.Allow("1.1.1.1", "t", "other@b.c") {
		t.Fatal("other email should allow")
	}
}
