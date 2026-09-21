// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package stt

import (
	"runtime"
	"sort"
	"sync"

	"github.com/shirou/gopsutil/v4/mem"
)

// Machine is what can cheaply be learned about the computer, used to rank
// models by whether they'll keep up.
type Machine struct {
	Cores    int
	MemoryMB int
}

// referenceCores matches the slowest machine the catalogue's realtime factors
// were measured on, so scaling from it errs toward caution.
const referenceCores = 8

var (
	machineOnce sync.Once
	machine     Machine
)

func Host() Machine {
	machineOnce.Do(func() {
		machine = Machine{Cores: runtime.NumCPU(), MemoryMB: totalMemoryMB()}
	})
	return machine
}

func totalMemoryMB() int {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return int(v.Total / (1 << 20))
}

// EstimatedRealtime scales the catalogue's measured factor by core count. A
// heuristic, not a benchmark: cores say nothing about clock speed or vector width.
func (m Model) EstimatedRealtime(host Machine) float64 {
	if m.RealtimeFactor <= 0 || host.Cores <= 0 {
		return 0
	}
	return m.RealtimeFactor * float64(host.Cores) / referenceCores
}

// Memory needed beyond the file itself. Loading maps the weights, then a
// batch run allocates compute buffers (KV cache, activations, the audio
// encoder's scratch) that scale with the model, roughly half its size again,
// plus a few hundred MB that every run needs regardless of size.
const (
	inferenceOverheadRatio = 1.5
	inferenceFixedMB       = 512
	// The rest of RAM belongs to the OS, the browser view and whatever else is
	// open; past this share the machine swaps and the "fast" label is a lie.
	memoryBudgetShare = 0.6
)

// FitsMemory checks the working set, not the download: the file is only the
// floor of what a transcription allocates.
func (m Model) FitsMemory(host Machine) bool {
	if host.MemoryMB <= 0 {
		return true
	}
	needed := m.SizeMB()*inferenceOverheadRatio + inferenceFixedMB
	return needed < float64(host.MemoryMB)*memoryBudgetShare
}

// Fit is how a model is expected to behave on this machine.
type Fit int

const (
	FitComfortable Fit = iota
	// No measured realtime factor (a user-dropped file); ranks below known-good
	// but "slow" would be a claim the catalogue can't support.
	FitUnknown
	FitSlow
	FitTooLarge
)

func (m Model) Fit(host Machine) Fit {
	switch {
	case !m.FitsMemory(host):
		return FitTooLarge
	case m.RealtimeFactor <= 0:
		return FitUnknown
	case m.EstimatedRealtime(host) < 2:
		return FitSlow
	default:
		return FitComfortable
	}
}

func (f Fit) String() string {
	switch f {
	case FitTooLarge:
		return "too-large"
	case FitSlow:
		return "slow"
	case FitUnknown:
		return "unknown"
	default:
		return "comfortable"
	}
}

// Label describes how a model is expected to keep up on this machine.
func (m Model) FitLabel(host Machine) string {
	switch m.Fit(host) {
	case FitTooLarge:
		return "May not fit in memory"
	case FitSlow:
		return "Slow on this machine"
	case FitUnknown:
		return "Speed unknown"
	default:
		if m.EstimatedRealtime(host) >= 10 {
			return "Very fast on this machine"
		}
		return "Fast on this machine"
	}
}

// RankForMachine puts what is already downloaded first, then what this machine
// can comfortably run, then the rest by accuracy.
func RankForMachine(models []Model, host Machine, downloaded func(Model) bool) {
	sort.SliceStable(models, func(i, j int) bool {
		a, b := models[i], models[j]

		if downloaded != nil {
			if has, other := downloaded(a), downloaded(b); has != other {
				return has
			}
		}
		if fa, fb := a.Fit(host), b.Fit(host); fa != fb {
			return fa < fb
		}
		if a.Recommended != b.Recommended {
			return a.Recommended
		}
		return a.AccuracyScore > b.AccuracyScore
	})
}

// Suggested is the model to propose for this machine regardless of what is
// downloaded: the most accurate one it runs comfortably, catalogue picks first.
func Suggested(models []Model, host Machine) (Model, bool) {
	var best Model
	found := false
	for _, model := range models {
		if model.Fit(host) != FitComfortable {
			continue
		}
		if !found || prefer(model, best) {
			best, found = model, true
		}
	}
	return best, found
}

func prefer(a, b Model) bool {
	if a.Recommended != b.Recommended {
		return a.Recommended
	}
	if a.AccuracyScore != b.AccuracyScore {
		return a.AccuracyScore > b.AccuracyScore
	}
	return a.SizeBytes < b.SizeBytes
}

// Recommended is the best downloaded model, else the best the machine can run.
func Recommended(models []Model, host Machine, downloaded func(Model) bool) (Model, bool) {
	ranked := append([]Model(nil), models...)
	RankForMachine(ranked, host, downloaded)
	if len(ranked) == 0 {
		return Model{}, false
	}
	return ranked[0], true
}
