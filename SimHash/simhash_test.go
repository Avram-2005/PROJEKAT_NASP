package SimHash

import "testing"

func TestHammingDistance_IdenticalTexts(t *testing.T) {
	text := "Zato sto su se carevi igrali rata!"
	dist := HammingDistance(text, text)

	if dist != 0 {
		t.Errorf("expected distance 0, got %d", dist)
	}
}

func TestHammingDistance_SimilarTexts(t *testing.T) {
	text1 := "Zato sto su se carevi igrali rata!"
	text2 := "Zato sto su se carevi igrali rata."

	dist := HammingDistance(text1, text2)

	if dist == 0 {
		t.Errorf("expected non-zero distance for similar texts")
	}
}

func TestHammingDistance_DifferentTexts(t *testing.T) {
	text1 := "Zato sto su se carevi igrali rata!"
	text2 := "Ja pevam balade i pesme lagane i cuvam svoju snagu za satre dobre dane"

	dist := HammingDistance(text1, text2)

	if dist <= 10 {
		t.Errorf("expected large distance for different texts, got %d", dist)
	}
}

func TestHammingDistance_EmptyStrings(t *testing.T) {
	dist := HammingDistance("", "")

	if dist != 0 {
		t.Errorf("expected distance 0 for empty strings, got %d", dist)
	}
}
