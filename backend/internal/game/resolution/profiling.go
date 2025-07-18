package resolution

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // Import pprof for profiling endpoints
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"
)

// ProfileManager handles profiling operations for the resolution engine
type ProfileManager struct {
	cpuProfileFile   *os.File
	memProfileFile   *os.File
	traceFile        *os.File
	profilingActive  bool
	profileStartTime time.Time
	mu               sync.RWMutex
}

// NewProfileManager creates a new profile manager
func NewProfileManager() *ProfileManager {
	return &ProfileManager{}
}

// StartCPUProfile starts CPU profiling to the specified file
func (pm *ProfileManager) StartCPUProfile(filename string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.cpuProfileFile != nil {
		return fmt.Errorf("CPU profiling already active")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}

	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return fmt.Errorf("failed to start CPU profiling: %w", err)
	}

	pm.cpuProfileFile = file
	pm.profilingActive = true
	pm.profileStartTime = time.Now()
	return nil
}

// StopCPUProfile stops CPU profiling
func (pm *ProfileManager) StopCPUProfile() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.cpuProfileFile == nil {
		return fmt.Errorf("CPU profiling not active")
	}

	pprof.StopCPUProfile()
	err := pm.cpuProfileFile.Close()
	pm.cpuProfileFile = nil

	if err != nil {
		return fmt.Errorf("failed to close CPU profile file: %w", err)
	}

	return nil
}

// WriteMemProfile writes a memory profile to the specified file
func (pm *ProfileManager) WriteMemProfile(filename string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create memory profile file: %w", err)
	}
	defer file.Close()

	// Force garbage collection to get accurate memory stats
	runtime.GC()

	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("failed to write memory profile: %w", err)
	}

	return nil
}

// StartTrace starts execution tracing to the specified file
func (pm *ProfileManager) StartTrace(filename string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.traceFile != nil {
		return fmt.Errorf("tracing already active")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create trace file: %w", err)
	}

	if err := trace.Start(file); err != nil {
		file.Close()
		return fmt.Errorf("failed to start tracing: %w", err)
	}

	pm.traceFile = file
	return nil
}

// StopTrace stops execution tracing
func (pm *ProfileManager) StopTrace() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.traceFile == nil {
		return fmt.Errorf("tracing not active")
	}

	trace.Stop()
	err := pm.traceFile.Close()
	pm.traceFile = nil

	if err != nil {
		return fmt.Errorf("failed to close trace file: %w", err)
	}

	return nil
}

// IsProfilingActive returns whether any profiling is currently active
func (pm *ProfileManager) IsProfilingActive() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.profilingActive || pm.traceFile != nil
}

// GetProfileDuration returns how long profiling has been active
func (pm *ProfileManager) GetProfileDuration() time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if !pm.profilingActive {
		return 0
	}

	return time.Since(pm.profileStartTime)
}

// ProfiledResolution runs a resolution with profiling enabled
func (pm *ProfileManager) ProfiledResolution(ctx context.Context, engine *DATCEngine, profilePrefix string) ([]AdjudicationResult, error) {
	// Start CPU profiling
	cpuFile := fmt.Sprintf("%s_cpu.prof", profilePrefix)
	if err := pm.StartCPUProfile(cpuFile); err != nil {
		return nil, fmt.Errorf("failed to start CPU profiling: %w", err)
	}

	// Start tracing
	traceFile := fmt.Sprintf("%s_trace.out", profilePrefix)
	if err := pm.StartTrace(traceFile); err != nil {
		pm.StopCPUProfile() // Clean up CPU profiling
		return nil, fmt.Errorf("failed to start tracing: %w", err)
	}

	// Run the resolution
	results := engine.ResolveAll()

	// Stop profiling
	if err := pm.StopCPUProfile(); err != nil {
		return results, fmt.Errorf("failed to stop CPU profiling: %w", err)
	}

	if err := pm.StopTrace(); err != nil {
		return results, fmt.Errorf("failed to stop tracing: %w", err)
	}

	// Write memory profile
	memFile := fmt.Sprintf("%s_mem.prof", profilePrefix)
	if err := pm.WriteMemProfile(memFile); err != nil {
		return results, fmt.Errorf("failed to write memory profile: %w", err)
	}

	return results, nil
}

// StartPProfServer starts an HTTP server with pprof endpoints
func (pm *ProfileManager) StartPProfServer(addr string) error {
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("pprof server error: %v\n", err)
		}
	}()

	fmt.Printf("pprof server started at http://%s/debug/pprof/\n", addr)
	return nil
}

// BenchmarkProfiler provides profiling capabilities specifically for benchmarks
type BenchmarkProfiler struct {
	manager *ProfileManager
	enabled bool
}

// NewBenchmarkProfiler creates a new benchmark profiler
func NewBenchmarkProfiler() *BenchmarkProfiler {
	return &BenchmarkProfiler{
		manager: NewProfileManager(),
		enabled: true,
	}
}

// Enable turns on benchmark profiling
func (bp *BenchmarkProfiler) Enable() {
	bp.enabled = true
}

// Disable turns off benchmark profiling
func (bp *BenchmarkProfiler) Disable() {
	bp.enabled = false
}

// ProfileBenchmark runs a benchmark with profiling
func (bp *BenchmarkProfiler) ProfileBenchmark(name string, fn func()) error {
	if !bp.enabled {
		fn()
		return nil
	}

	timestamp := time.Now().Format("20060102_150405")
	profilePrefix := fmt.Sprintf("benchmark_%s_%s", name, timestamp)

	// Start CPU profiling
	cpuFile := fmt.Sprintf("%s_cpu.prof", profilePrefix)
	if err := bp.manager.StartCPUProfile(cpuFile); err != nil {
		return fmt.Errorf("failed to start CPU profiling: %w", err)
	}

	// Run the benchmark
	fn()

	// Stop CPU profiling
	if err := bp.manager.StopCPUProfile(); err != nil {
		return fmt.Errorf("failed to stop CPU profiling: %w", err)
	}

	// Write memory profile
	memFile := fmt.Sprintf("%s_mem.prof", profilePrefix)
	if err := bp.manager.WriteMemProfile(memFile); err != nil {
		return fmt.Errorf("failed to write memory profile: %w", err)
	}

	fmt.Printf("Benchmark %s profiled: CPU=%s, Memory=%s\n", name, cpuFile, memFile)
	return nil
}

// Global profile manager
var globalProfileManager = NewProfileManager()

// GetGlobalProfileManager returns the global profile manager
func GetGlobalProfileManager() *ProfileManager {
	return globalProfileManager
}

// StartGlobalCPUProfile starts global CPU profiling
func StartGlobalCPUProfile(filename string) error {
	return globalProfileManager.StartCPUProfile(filename)
}

// StopGlobalCPUProfile stops global CPU profiling
func StopGlobalCPUProfile() error {
	return globalProfileManager.StopCPUProfile()
}

// WriteGlobalMemProfile writes a global memory profile
func WriteGlobalMemProfile(filename string) error {
	return globalProfileManager.WriteMemProfile(filename)
}

// StartGlobalPProfServer starts the global pprof server
func StartGlobalPProfServer(addr string) error {
	return globalProfileManager.StartPProfServer(addr)
}
