package vreme

import (
	"io"
	"math"
	"os"
	"strings"
	"testing"
)

// ponastaviNastavitve sets the global settings for one test and
// restores the old values when the test finishes.
func ponastaviNastavitve(t *testing.T, minT, maxT float64, st int) {
	t.Helper()
	stariMin, stariMax, stariSt := MinTemp, MaxTemp, StMeritev
	MinTemp, MaxTemp, StMeritev = minT, maxT, st
	t.Cleanup(func() {
		MinTemp, MaxTemp, StMeritev = stariMin, stariMax, stariSt
	})
}

// priblizno compares floats with a small tolerance (averages are not always exact).
func priblizno(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// zajemiIzpis runs f and returns everything it printed to stdout.
func zajemiIzpis(t *testing.T, f func()) string {
	t.Helper()
	stari := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = stari
	izpis, _ := io.ReadAll(r)
	return string(izpis)
}

// ---------- DodajPostajo ----------

func TestDodajPostajo(t *testing.T) {
	postaje := map[string]Postaja{}

	if err := DodajPostajo("P1", "Ljubljana", "lj1", postaje); err != nil {
		t.Fatalf("nepričakovana napaka: %v", err)
	}

	p, obstaja := postaje["lj1"]
	if !obstaja {
		t.Fatal("postaja lj1 ni bila dodana v mapo")
	}
	if p.ime != "P1" || p.kraj != "Ljubljana" {
		t.Errorf("napačni podatki: ime=%q kraj=%q", p.ime, p.kraj)
	}
	if len(p.meritve) != 0 {
		t.Errorf("nova postaja naj nima meritev, ima %d", len(p.meritve))
	}
}

func TestDodajPostajoPodvojenKljuc(t *testing.T) {
	postaje := map[string]Postaja{}
	DodajPostajo("P1", "Ljubljana", "lj1", postaje)

	err := DodajPostajo("P2", "Maribor", "lj1", postaje)
	if err == nil {
		t.Fatal("pričakovana napaka pri podvojenem ključu")
	}
	// the original station must not be overwritten
	if postaje["lj1"].kraj != "Ljubljana" {
		t.Errorf("obstoječa postaja je bila prepisana: kraj=%q", postaje["lj1"].kraj)
	}
}

func TestDodajVecPostaj(t *testing.T) {
	postaje := map[string]Postaja{}
	DodajPostajo("P1", "Ljubljana", "lj1", postaje)
	DodajPostajo("P2", "Maribor", "mb1", postaje)
	DodajPostajo("P3", "Koper", "kp1", postaje)

	if len(postaje) != 3 {
		t.Errorf("pričakovane 3 postaje, dobil %d", len(postaje))
	}
}

// ---------- DodajMeritev ----------

func TestDodajMeritev(t *testing.T) {
	ponastaviNastavitve(t, -50, 50, 3)
	postaje := map[string]Postaja{}
	DodajPostajo("P1", "Ljubljana", "lj1", postaje)

	for _, m := range []float64{4.5, -3, 12} {
		if err := DodajMeritev(postaje, "lj1", m); err != nil {
			t.Fatalf("nepričakovana napaka pri %.1f: %v", m, err)
		}
	}

	got := postaje["lj1"].meritve
	want := []float64{4.5, -3, 12}
	if len(got) != len(want) {
		t.Fatalf("pričakovane %d meritve, dobil %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("meritev[%d] = %v, pričakovano %v", i, got[i], want[i])
		}
	}
}

func TestDodajMeritevNeobstojecaPostaja(t *testing.T) {
	ponastaviNastavitve(t, -50, 50, 3)
	postaje := map[string]Postaja{}

	if err := DodajMeritev(postaje, "ne_obstaja", 10); err == nil {
		t.Fatal("pričakovana napaka za neobstoječo postajo")
	}
	if len(postaje) != 0 {
		t.Error("DodajMeritev ne sme ustvariti nove postaje")
	}
}

// TestDodajMeritevObmocje checks the accepted temperature range with a
// table-driven test, one row per case.
func TestDodajMeritevObmocje(t *testing.T) {
	primeri := []struct {
		ime         string
		minT, maxT  float64
		temperatura float64
		napaka      bool
	}{
		{"privzeto, sredina", -50, 50, 20, false},
		{"privzeto, točno min", -50, 50, -50, false},
		{"privzeto, točno max", -50, 50, 50, false},
		{"privzeto, pod min", -50, 50, -50.1, true},
		{"privzeto, nad max", -50, 50, 50.1, true},
		{"ožje območje, v redu", -10, 30, 25, false},
		{"ožje območje, prevroče", -10, 30, 31, true},
		{"ožje območje, premrzlo", -10, 30, -11, true},
		{"širše območje, 120 dovoljeno", -100, 150, 120, false},
	}

	for _, p := range primeri {
		t.Run(p.ime, func(t *testing.T) {
			ponastaviNastavitve(t, p.minT, p.maxT, 3)
			postaje := map[string]Postaja{}
			DodajPostajo("P1", "Ljubljana", "lj1", postaje)

			err := DodajMeritev(postaje, "lj1", p.temperatura)

			if p.napaka && err == nil {
				t.Errorf("%.1f v [%.1f, %.1f]: pričakovana napaka", p.temperatura, p.minT, p.maxT)
			}
			if !p.napaka && err != nil {
				t.Errorf("%.1f v [%.1f, %.1f]: nepričakovana napaka: %v", p.temperatura, p.minT, p.maxT, err)
			}
			// a rejected measurement must not be stored
			if p.napaka && len(postaje["lj1"].meritve) != 0 {
				t.Error("zavrnjena meritev je bila vseeno shranjena")
			}
		})
	}
}

// ---------- statistika ----------

func TestStatistika(t *testing.T) {
	primeri := []struct {
		ime       string
		stMeritev int
		maxT      float64
		meritve   []float64
		ok        bool
		najnizja  float64
		najvisja  float64
		povprecna float64
	}{
		{"tri meritve", 3, 50, []float64{10, 20, 30}, true, 10, 30, 20},
		{"negativne", 3, 50, []float64{-5, -15, -10}, true, -15, -5, -10},
		{"mešane", 3, 50, []float64{-4, 0, 4, 8}, true, -4, 8, 2},
		{"vse enake", 3, 50, []float64{7, 7, 7}, true, 7, 7, 7},
		{"decimalke", 3, 50, []float64{1.5, 2.5, 3.5}, true, 1.5, 3.5, 2.5},
		{"premalo meritev", 3, 50, []float64{10, 20}, false, 0, 0, 0},
		{"nič meritev", 3, 50, nil, false, 0, 0, 0},
		{"stMeritev=1, ena meritev", 1, 50, []float64{12}, true, 12, 12, 12},
		{"stMeritev=5, štiri meritve", 5, 50, []float64{1, 2, 3, 4}, false, 0, 0, 0},
		// catches the "najnizja = 100" starting-value bug
		{"vse nad 100", 3, 200, []float64{110, 120, 130}, true, 110, 130, 120},
		// catches the "najvisja = -100" starting-value bug
		{"vse pod -100", 3, 200, []float64{-110, -120, -130}, true, -130, -110, -120},
	}

	for _, p := range primeri {
		t.Run(p.ime, func(t *testing.T) {
			ponastaviNastavitve(t, -200, p.maxT, p.stMeritev)
			postaje := map[string]Postaja{}
			DodajPostajo("P1", "Ljubljana", "lj1", postaje)
			for _, m := range p.meritve {
				if err := DodajMeritev(postaje, "lj1", m); err != nil {
					t.Fatalf("priprava: DodajMeritev(%.1f): %v", m, err)
				}
			}

			najnizja, najvisja, povprecna, ok := statistika(postaje, "lj1")

			if ok != p.ok {
				t.Fatalf("ok = %v, pričakovano %v", ok, p.ok)
			}
			if !ok {
				return
			}
			if !priblizno(najnizja, p.najnizja) {
				t.Errorf("najnizja = %v, pričakovano %v", najnizja, p.najnizja)
			}
			if !priblizno(najvisja, p.najvisja) {
				t.Errorf("najvisja = %v, pričakovano %v", najvisja, p.najvisja)
			}
			if !priblizno(povprecna, p.povprecna) {
				t.Errorf("povprecna = %v, pričakovano %v", povprecna, p.povprecna)
			}
		})
	}
}

func TestStatistikaNeobstojecaPostaja(t *testing.T) {
	ponastaviNastavitve(t, -50, 50, 3)
	postaje := map[string]Postaja{}

	_, _, _, ok := statistika(postaje, "ne_obstaja")
	if ok {
		t.Error("statistika za neobstoječo postajo mora vrniti ok=false")
	}
}

// ---------- IzpisPorocila (checks printed output) ----------

func TestIzpisPorocilaOpis(t *testing.T) {
	primeri := []struct {
		meritve []float64
		opis    string
	}{
		{[]float64{30, 30, 30}, "vroče"},
		{[]float64{25, 25, 25}, "vroče"}, // boundary: exactly 25
		{[]float64{20, 20, 20}, "toplo"},
		{[]float64{15, 15, 15}, "toplo"}, // boundary: exactly 15
		{[]float64{10, 10, 10}, "hladno"},
		{[]float64{5, 5, 5}, "hladno"}, // boundary: exactly 5
		{[]float64{0, 0, 0}, "mrzlo"},
		{[]float64{-10, -10, -10}, "mrzlo"},
	}

	for _, p := range primeri {
		t.Run(p.opis, func(t *testing.T) {
			ponastaviNastavitve(t, -50, 50, 3)
			postaje := map[string]Postaja{}
			DodajPostajo("P1", "Ljubljana", "lj1", postaje)
			for _, m := range p.meritve {
				DodajMeritev(postaje, "lj1", m)
			}

			izpis := strings.ToLower(zajemiIzpis(t, func() { IzpisPorocila(postaje) }))

			if !strings.Contains(izpis, p.opis) {
				t.Errorf("povprečje %.1f: pričakovan opis %q, izpis:\n%s", p.meritve[0], p.opis, izpis)
			}
		})
	}
}

func TestIzpisPorocilaPremaloMeritev(t *testing.T) {
	ponastaviNastavitve(t, -50, 50, 3)
	postaje := map[string]Postaja{}
	DodajPostajo("P1", "Ljubljana", "lj1", postaje)
	DodajMeritev(postaje, "lj1", 10)

	izpis := zajemiIzpis(t, func() { IzpisPorocila(postaje) })

	if strings.Contains(izpis, "max") {
		t.Errorf("statistika se ne sme izpisati pri premalo meritvah, izpis:\n%s", izpis)
	}
	if !strings.Contains(izpis, "lj1") {
		t.Errorf("izpis naj omeni postajo lj1, izpis:\n%s", izpis)
	}
}