package mode

import "testing"

func TestNormalizeID(t *testing.T) {
	tests := map[string]string{
		"":           Solo1v1,
		"1v1":        Solo1v1,
		"2v2":        Solo2v2,
		"cross_2v2":  Solo2v2,
		"3p_chain":   Solo3pChain,
		"杀上保下":       Solo3pChain,
		"3p_ddz":     Solo3pDdz,
		"斗地主":        Solo3pDdz,
		"unknown":    Solo1v1,
	}
	for in, want := range tests {
		if got := NormalizeID(in); got != want {
			t.Fatalf("NormalizeID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRegistryLookup(t *testing.T) {
	m, ok := Lookup(Solo2v2)
	if !ok || m.LayoutKey != LayoutCross2v2 {
		t.Fatalf("lookup 2v2: ok=%v meta=%+v", ok, m)
	}
}

func TestRegistryAll(t *testing.T) {
	all := All()
	if len(all) < 4 {
		t.Fatalf("expected at least 4 modes, got %d", len(all))
	}
}
