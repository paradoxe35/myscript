// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package stt

import (
	"runtime"
	"sort"
	"sync"

	"github.com/shirou/gopsutil/v4/mem"
)

type Machine struct {
	Cores    int
	MemoryMB int
}

// The slowest machine the catalogue's realtime factors were measured on.
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

// A heuristic: cores say nothing about clock speed or vector width.
func (m Model) EstimatedRealtime(host Machine) float64 {
	if m.RealtimeFactor <= 0 || host.Cores <= 0 {
		return 0
	}
	return m.RealtimeFactor * float64(host.Cores) / referenceCores
}

// Beyond the file itself, a run allocates compute buffers of roughly half the
// model's size plus a few hundred MB regardless of size.
const (
	inferenceOverheadRatio = 1.5
	inferenceFixedMB       = 512
	// The rest belongs to the OS and the browser view; past this share the
	// machine swaps.
	memoryBudgetShare = 0.6
)

// Checks the working set, not the download.
func (m Model) FitsMemory(host Machine) bool {
	if host.MemoryMB <= 0 {
		return true
	}
	needed := m.SizeMB()*inferenceOverheadRatio + inferenceFixedMB
	return needed < float64(host.MemoryMB)*memoryBudgetShare
}

type Fit int

const (
	FitComfortable Fit = iota
	// No measured realtime factor (a user-dropped file); "slow" would be a
	// claim the catalogue cannot support.
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

// Downloaded first, then what runs comfortably, then by accuracy.
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

// The most accurate model that runs comfortably, regardless of downloads.
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

// The best downloaded model, else the best the machine can run.
func Recommended(models []Model, host Machine, downloaded func(Model) bool) (Model, bool) {
	ranked := append([]Model(nil), models...)
	RankForMachine(ranked, host, downloaded)
	if len(ranked) == 0 {
		return Model{}, false
	}
	return ranked[0], true
}
