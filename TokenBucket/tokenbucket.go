package TokenBucket

import (
	"encoding/binary"
	"fmt"
	"time"
)

const INTERNAL_KEY = "__tb__state"

type TokenBucket struct {
	maxNumTokens   int64
	currentTokens  int64
	refillTime     int64
	lastTimeRefill time.Time
}

func NewTokenBucket(maxNumTokens int64, refillInterval int64) (*TokenBucket, error) {
	if maxNumTokens <= 0 {
		return nil, fmt.Errorf("Maximum number of tokens must be greater than 0")
	}
	if refillInterval <= 0 {
		return nil, fmt.Errorf("Interval for refill must be greater than 0")
	}
	return &TokenBucket{
		maxNumTokens:   maxNumTokens,
		currentTokens:  maxNumTokens,
		refillTime:     refillInterval,
		lastTimeRefill: time.Now(),
	}, nil
}

func (tb *TokenBucket) GetNextRefill() int64 {
	now := time.Now()
	deltat := now.Sub(tb.lastTimeRefill)
	deltatSeconds := deltat.Seconds()
	timeToNextRefill := tb.refillTime - int64(deltatSeconds)
	return timeToNextRefill
}

// puni bucket na osnovu proteklog vremena od poslednjeg punjenja
func (tb *TokenBucket) refill() {
	now := time.Now()
	deltat := now.Sub(tb.lastTimeRefill)
	deltatSeconds := deltat.Seconds()
	intervals := int64(deltatSeconds / float64(tb.refillTime)) //br intrevala koji je prosao
	if intervals <= 0 {
		return
	}
	tb.currentTokens += intervals * tb.maxNumTokens
	if tb.currentTokens > tb.maxNumTokens {
		tb.currentTokens = tb.maxNumTokens
	}
	tb.lastTimeRefill = tb.lastTimeRefill.Add(time.Duration(intervals*tb.refillTime) * time.Second)
}

// provera da li je zahtev odobren, trosi jedan token
func (tb *TokenBucket) Allow() bool {
	tb.refill()
	if tb.currentTokens <= 0 {
		return false
	}
	tb.currentTokens--
	return true
}

// serijalizuje token bucket
// maxNumTokens 8B | currentTokens 8B | refillTime 8B | lastTimeRefill 8B
func (tb *TokenBucket) Serialize() []byte {
	buf := make([]byte, 32)
	binary.BigEndian.PutUint64(buf[0:8], uint64(tb.maxNumTokens))
	binary.BigEndian.PutUint64(buf[8:16], uint64(tb.currentTokens))
	binary.BigEndian.PutUint64(buf[16:24], uint64(tb.refillTime))
	binary.BigEndian.PutUint64(buf[24:32], uint64(tb.lastTimeRefill.UnixNano()))
	return buf
}

func Deserialize(data []byte) (*TokenBucket, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("Not enough data for deserialization of token bucket")
	}
	maxNumTokens := int64(binary.BigEndian.Uint64(data[0:8]))
	currentTokens := int64(binary.BigEndian.Uint64(data[8:16]))
	refillInterval := int64(binary.BigEndian.Uint64(data[16:24]))
	lastTimeRefill := time.Unix(0, int64(binary.BigEndian.Uint64(data[24:32])))
	return &TokenBucket{
		maxNumTokens:   maxNumTokens,
		currentTokens:  currentTokens,
		refillTime:     refillInterval,
		lastTimeRefill: lastTimeRefill,
	}, nil
}

func (tb *TokenBucket) CurrentTokens() int64 {
	return tb.currentTokens
}

func (tb *TokenBucket) MaxNumTokens() int64 {
	return tb.maxNumTokens
}
