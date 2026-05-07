package record

import (
	"testing"
	"time"
)

func TestRecordSerialization(t *testing.T) {
	original, err := NewRecord("testKey", []byte("testValue"), false, time.Now())
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	serialized := original.Serialize()
	deserialized, _, err := DeserializeRecord(serialized)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	if original.Timestamp.UnixNano() != deserialized.Timestamp.UnixNano() {
		t.Errorf("Timestamp mismatch: expected %v, got %v", original.Timestamp, deserialized.Timestamp)
	}
	if original.Tombstone != deserialized.Tombstone {
		t.Errorf("Tombstone mismatch: expected %v, got %v", original.Tombstone, deserialized.Tombstone)
	}
	if original.Key != deserialized.Key {
		t.Errorf("Key mismatch: expected %s, got %s", original.Key, deserialized.Key)
	}
	if string(original.Value) != string(deserialized.Value) {
		t.Errorf("Value mismatch: expected %s, got %s", original.Value, deserialized.Value)
	}
}

func TestRecordDeserializationWithCorruptedData(t *testing.T) {
	original, err := NewRecord("testKey", []byte("testValue"), false, time.Now())
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	serialized := original.Serialize()
	// Corrupt the data by changing a byte
	serialized[10] = 0xFF

	_, _, err = DeserializeRecord(serialized)
	if err == nil {
		t.Fatal("Expected deserialization to fail due to CRC mismatch, but it succeeded")
	}
}
