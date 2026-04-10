package TokenBucket

import (
	"testing"
	"time"
)

// konstruktor
func TestNewTokenBucket(t *testing.T) {
	tb, err := NewTokenBucket(10, time.Second)
	if err != nil {
		t.Fatalf("Expected succsessful creation of token bucket, but got: %v", err)
	}
	if tb.CurrentTokens() != 10 {
		t.Fatalf("Expected 10 tokens at the start, but got %d instead", tb.CurrentTokens())
	}
}

func TestNewTokenBucketInvalidMaxNumTokens(t *testing.T) {
	_, err := NewTokenBucket(0, time.Second)
	if err == nil {
		t.Fatal("Expected an error for invalid number of maximum tokens")
	}
	_, err = NewTokenBucket(-5, time.Second)
	if err == nil {
		t.Fatal("Expected error for negative number of max tokens")
	}
}

func TestNewTokenBucketInvalidInterval(t *testing.T) {
	_, err := NewTokenBucket(10, 0)
	if err == nil {
		t.Fatal("Expected an error for invalid interval")
	}
}

// Allow
func TestAllow(t *testing.T) {
	tb, _ := NewTokenBucket(5, time.Second)
	if !tb.Allow() {
		t.Fatal("First request must be allowed")
	}
	if tb.CurrentTokens() != 4 {
		t.Fatalf("Expected 4 tokens after one request, got %d instead", tb.CurrentTokens())
	}
}

func TestAllowZeroOutput(t *testing.T) {
	tb, _ := NewTokenBucket(3, time.Hour) //veliki interval kako za vreme testa ne bi refill uradio
	tb.Allow()
	tb.Allow()
	tb.Allow()
	if tb.CurrentTokens() != 0 {
		t.Fatalf("Expected 0 tokens, gpt %d instead", tb.CurrentTokens())
	}
}
