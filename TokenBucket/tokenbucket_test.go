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

func TestAllowEmpty(t *testing.T) {
	tb, _ := NewTokenBucket(2, time.Hour)
	tb.Allow()
	tb.Allow()
	if tb.Allow() {
		t.Fatal("Request must be rejected when there is no token")
	}
}

func TestAllowAfterRefill(t *testing.T) {
	tb, _ := NewTokenBucket(2, 50*time.Millisecond)
	tb.Allow()
	tb.Allow()
	if tb.Allow() {
		t.Fatal("Request must be rejected when there are no tokens")
	}
	time.Sleep(60 * time.Millisecond) //pauza jedan interval
	if !tb.Allow() {
		t.Fatal("Request must be allowed after refill")
	}
}

// refill
func TestRefillLessThanMax(t *testing.T) {
	tb, _ := NewTokenBucket(5, 50*time.Millisecond)
	time.Sleep(200 * time.Millisecond) //pauza vise intervala
	tb.Allow()                         //pokrece refill
	if tb.CurrentTokens() > tb.MaxNumTokens() {
		t.Fatalf("Token Bucket must not exceed maximum capacity")
	}

}

func TestRefillMultipleIntervals(t *testing.T) {
	tb, _ := NewTokenBucket(3, 30*time.Millisecond)
	tb.Allow()
	tb.Allow()
	tb.Allow()
	time.Sleep(70 * time.Millisecond) //pauza dva intervala
	tb.refill()
	if tb.CurrentTokens() != 3 {
		t.Fatalf("Expected 3 tokens after refill, got %d", tb.CurrentTokens())
	}
}

// serijalizacija
func TestSerializeDeserialize(t *testing.T) {
	tb, _ := NewTokenBucket(10, time.Second)
	tb.Allow()
	tb.Allow()
	data := tb.Serialize()
	if len(data) != 32 {
		t.Fatalf("Expected 32B but got %d", len(data))
	}
	tb2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Failed deserialization: %v", err)
	}
	if tb2.maxNumTokens != tb.maxNumTokens {
		t.Fatal("Max number of tokens are not equal")
	}
	if tb2.currentTokens != tb.currentTokens {
		t.Fatal("current number of tokens are not equal")
	}
	if tb2.refillTime != tb.refillTime {
		t.Fatal("Refill Time of tokens are not equal")
	}
}

func TestDeserializeInvalidData(t *testing.T) {
	_, err := Deserialize([]byte{1, 2, 3}) //izraz koji nije kompletan
	if err == nil {
		t.Fatal("Expected error for invalid input")
	}
}

func TestDeserializePreserveState(t *testing.T) {
	tb, _ := NewTokenBucket(5, time.Minute)
	tb.Allow()
	tb.Allow()
	tb.Allow()
	data := tb.Serialize()
	tb2, _ := Deserialize(data)
	if tb2.CurrentTokens() != 2 {
		t.Fatalf("Expected two tokens after deserialization, but got %d instead", tb.CurrentTokens())
	}
	if !tb2.Allow() {
		t.Fatal("Request must be allowed after deserialization")
	}
}

func TestInternalKey(t *testing.T) {
	if len(INTERNAL_KEY) < 4 || INTERNAL_KEY[:4] != "__tb" {
		t.Fatal("Internal key must start with '__tb'")
	}
}
