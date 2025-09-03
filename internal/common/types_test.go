package common

import "testing"

func TestArrayTypeStat_TypeDisplay_SortedKeys(t *testing.T) {
	arr := NewArrayTypeStat()
	arr.Items["zeta"] = 1
	arr.Items["alpha"] = 2
	arr.Items["gamma"] = 3

	got := arr.TypeDisplay()
	want := "array[alpha, gamma, zeta]"
	if got != want {
		t.Fatalf("unexpected TypeDisplay: got %q want %q", got, want)
	}
}

func TestArrayTypeStat_TypeDisplay_Empty(t *testing.T) {
	arr := NewArrayTypeStat()
	got := arr.TypeDisplay()
	want := "array"
	if got != want {
		t.Fatalf("unexpected TypeDisplay for empty: got %q want %q", got, want)
	}
}
