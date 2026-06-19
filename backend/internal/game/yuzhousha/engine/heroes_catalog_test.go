package engine

import "testing"

func TestListHeroesPagination(t *testing.T) {
	page1 := ListHeroes(HeroesQuery{Page: 1, PageSize: 10})
	if len(page1.Heroes) != 10 {
		t.Fatalf("page1 len=%d want 10", len(page1.Heroes))
	}
	expectedPages := (page1.Total + 9) / 10 // 向上取整
	if page1.TotalPages != expectedPages {
		t.Fatalf("total=%d pages=%d want %d", page1.Total, page1.TotalPages, expectedPages)
	}
	
	pageLast := ListHeroes(HeroesQuery{Page: page1.TotalPages, PageSize: 10})
	expectedLastLen := page1.Total % 10
	if expectedLastLen == 0 {
		expectedLastLen = 10
	}
	if len(pageLast.Heroes) != expectedLastLen {
		t.Fatalf("last page len=%d want %d", len(pageLast.Heroes), expectedLastLen)
	}
}

func TestListHeroesKingdomFilter(t *testing.T) {
	shu := ListHeroes(HeroesQuery{Kingdom: "shu", PageSize: 50})
	if shu.Total != 7 {
		t.Fatalf("shu total=%d want 7", shu.Total)
	}
	for _, h := range shu.Heroes {
		if h.Kingdom != "shu" {
			t.Fatalf("unexpected kingdom %s for %s", h.Kingdom, h.ID)
		}
		if h.AccentColor == "" {
			t.Fatalf("hero %s missing accent_color", h.ID)
		}
		if h.DefaultSkinID == "" {
			t.Fatalf("hero %s missing default_skin_id", h.ID)
		}
	}
}

func TestValidateHeroForModeStandardPool(t *testing.T) {
	if err := ValidateHeroForMode("1v1", "liu_bei"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHeroForMode("2v2", "guan_yu"); err != nil {
		t.Fatal(err)
	}
}
