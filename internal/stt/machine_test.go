package stt

import "testing"

func TestEstimatedRealtimeScalesWithCores(t *testing.T) {
	model := Model{RealtimeFactor: 8}

	weak := model.EstimatedRealtime(Machine{Cores: 4})
	same := model.EstimatedRealtime(Machine{Cores: referenceCores})
	strong := model.EstimatedRealtime(Machine{Cores: 16})

	if same != 8 {
		t.Errorf("reference machine should reproduce the measured factor, got %v", same)
	}
	if !(weak < same && same < strong) {
		t.Errorf("expected %v < %v < %v", weak, same, strong)
	}
	if got := (Model{}).EstimatedRealtime(Machine{Cores: 8}); got != 0 {
		t.Errorf("an unmeasured model should not claim a speed, got %v", got)
	}
}

func TestFitsMemoryLeavesHeadroom(t *testing.T) {
	host := Machine{Cores: 8, MemoryMB: 8000}

	if !(Model{SizeBytes: 1 << 30}).FitsMemory(host) {
		t.Error("1 GB should fit in 8 GB")
	}
	if (Model{SizeBytes: 6 << 30}).FitsMemory(host) {
		t.Error("6 GB should not be offered on an 8 GB machine")
	}
	if !(Model{SizeBytes: 40 << 30}).FitsMemory(Machine{Cores: 8}) {
		t.Error("with memory unknown the check should pass rather than exclude")
	}
}

func TestFitClassification(t *testing.T) {
	host := Machine{Cores: 8, MemoryMB: 8000}

	cases := []struct {
		name  string
		model Model
		want  Fit
		label string
	}{
		{"fast and small", Model{RealtimeFactor: 20, SizeBytes: 100 << 20}, FitComfortable, "Very fast on this machine"},
		{"fast enough", Model{RealtimeFactor: 4, SizeBytes: 100 << 20}, FitComfortable, "Fast on this machine"},
		{"slow", Model{RealtimeFactor: 1, SizeBytes: 100 << 20}, FitSlow, "Slow on this machine"},
		{"huge", Model{RealtimeFactor: 20, SizeBytes: 7 << 30}, FitTooLarge, "May not fit in memory"},
		{"unmeasured", Model{SizeBytes: 100 << 20}, FitUnknown, "Speed unknown"},
	}
	for _, c := range cases {
		if got := c.model.Fit(host); got != c.want {
			t.Errorf("%s: Fit() = %v, want %v", c.name, got, c.want)
		}
		if got := c.model.FitLabel(host); got != c.label {
			t.Errorf("%s: FitLabel() = %q, want %q", c.name, got, c.label)
		}
	}
}

func TestRankPrefersDownloadedThenFit(t *testing.T) {
	host := Machine{Cores: 8, MemoryMB: 8000}

	huge := Model{ID: "huge", RealtimeFactor: 20, SizeBytes: 7 << 30}
	slow := Model{ID: "slow", RealtimeFactor: 1, SizeBytes: 100 << 20}
	quick := Model{ID: "quick", RealtimeFactor: 20, SizeBytes: 100 << 20}

	models := []Model{huge, slow, quick}
	RankForMachine(models, host, nil)
	if models[0].ID != "quick" || models[1].ID != "slow" || models[2].ID != "huge" {
		t.Fatalf("ranked %s, %s, %s; want quick, slow, huge", models[0].ID, models[1].ID, models[2].ID)
	}

	models = []Model{quick, slow}
	RankForMachine(models, host, func(m Model) bool { return m.ID == "slow" })
	if models[0].ID != "slow" {
		t.Errorf("a downloaded model should sort first, got %s", models[0].ID)
	}

	best, ok := Recommended([]Model{huge, slow, quick}, host, nil)
	if !ok || best.ID != "quick" {
		t.Errorf("Recommended = %s, want quick", best.ID)
	}
}

func TestHostReportsCores(t *testing.T) {
	if Host().Cores <= 0 {
		t.Error("core count should be positive")
	}
}

func TestSuggestedPrefersAccurateModelsTheMachineRunsWell(t *testing.T) {
	host := Machine{Cores: 8, MemoryMB: 16000}
	models := []Model{
		{ID: "huge", SizeBytes: 10 << 30, RealtimeFactor: 20, AccuracyScore: 0.99, Recommended: true},
		{ID: "slow", SizeBytes: 1 << 30, RealtimeFactor: 1, AccuracyScore: 0.95, Recommended: true},
		{ID: "plain", SizeBytes: 1 << 30, RealtimeFactor: 8, AccuracyScore: 0.9},
		{ID: "pick", SizeBytes: 1 << 30, RealtimeFactor: 8, AccuracyScore: 0.8, Recommended: true},
		{ID: "custom", SizeBytes: 1 << 30},
	}

	got, ok := Suggested(models, host)
	if !ok || got.ID != "pick" {
		t.Fatalf("suggested %q, want the recommended model that runs comfortably", got.ID)
	}

	if _, ok := Suggested([]Model{{ID: "custom", SizeBytes: 1 << 30}}, host); ok {
		t.Error("a model with no measured speed must not be suggested")
	}
}
