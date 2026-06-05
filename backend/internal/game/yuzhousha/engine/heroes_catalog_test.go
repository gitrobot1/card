package engine

import "testing"

func TestListHeroesPagination(t *testing.T) {
	page1 := ListHeroes(HeroesQuery{Page: 1, PageSize: 10})
	if len(page1.Heroes) != 10 {
		t.Fatalf("page1 len=%d want 10", len(page1.Heroes))
	}
	if page1.Total != 34 || page1.TotalPages != 4 {
		t.Fatalf("total=%d pages=%d want 34/4", page1.Total, page1.TotalPages)
	}

	page4 := ListHeroes(HeroesQuery{Page: 4, PageSize: 10})
	if len(page4.Heroes) != 4 {
		t.Fatalf("page4 len=%d want 4", len(page4.Heroes))
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
