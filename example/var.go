// file: variable_center_demo.go
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ---------------------------
// Types / Meta
// ---------------------------

// VariableType is descriptive only (no strict runtime enforcement here)
type VariableType string

const (
	TypeString VariableType = "string"
	TypeInt    VariableType = "int"
	TypeFloat  VariableType = "float"
	TypeBool   VariableType = "bool"
	TypeMap    VariableType = "map"
	TypeAny    VariableType = "any"
)

// StopType for future extension (not used directly here)
type StopType int

// VariableMeta describes a variable
type VariableMeta struct {
	Key         string       // e.g. "user.age" or "device.riskScore"
	Name        string       // human readable
	Type        VariableType // type hint
	Category    string       // domain: "user","device",...
	Source      string       // textual source description
	FetcherName string       // which fetcher to use if any
	ComputeName string       // which compute function to use (optional)
	Depends     []string     // dependent variable keys
	Cached      bool         // whether we can cache result in request cache
	TTLSeconds  int          // per-request TTL (0 = no expiry within request)
	Version     string
	Desc        string
}

// ---------------------------
// Context (per request)
// ---------------------------

type VarContext struct {
	Input map[string]any // raw input (from request)
	// Per-request cache and meta
	cacheMu sync.RWMutex
	cache   map[string]cacheValue
	// trace of accessed variables (for explainability)
	Trace []string
	// visiting set for cycle detection
	visiting map[string]bool
}

type cacheValue struct {
	val       any
	timestamp time.Time
	ttl       int // seconds
}

// NewVarContext creates a new context for a request
func NewVarContext(input map[string]any) *VarContext {
	return &VarContext{
		Input:    input,
		cache:    make(map[string]cacheValue),
		Trace:    make([]string, 0, 32),
		visiting: make(map[string]bool),
	}
}

// get from cache (thread-safe)
func (vc *VarContext) getCached(key string) (any, bool) {
	vc.cacheMu.RLock()
	defer vc.cacheMu.RUnlock()
	cv, ok := vc.cache[key]
	if !ok {
		return nil, false
	}
	if cv.ttl > 0 {
		if time.Since(cv.timestamp) > time.Duration(cv.ttl)*time.Second {
			// expired
			return nil, false
		}
	}
	return cv.val, true
}

// set cache
func (vc *VarContext) setCache(key string, val any, ttl int) {
	vc.cacheMu.Lock()
	defer vc.cacheMu.Unlock()
	vc.cache[key] = cacheValue{val: val, timestamp: time.Now(), ttl: ttl}
}

// add trace
func (vc *VarContext) addTrace(key string) {
	vc.Trace = append(vc.Trace, key)
}

// ---------------------------
// Fetcher / Compute function interfaces
// ---------------------------

type FetcherFunc func(ctx context.Context, vc *VarContext, meta VariableMeta) (any, error)
type ComputeFunc func(ctx context.Context, vc *VarContext, meta VariableMeta, vcCenter *VariableCenter) (any, error)

// ---------------------------
// VariableCenter
// ---------------------------

type VariableCenter struct {
	metasMu sync.RWMutex
	metas   map[string]VariableMeta

	fetchersMu sync.RWMutex
	fetchers   map[string]FetcherFunc

	computesMu sync.RWMutex
	computes   map[string]ComputeFunc
}

func NewVariableCenter() *VariableCenter {
	return &VariableCenter{
		metas:    make(map[string]VariableMeta),
		fetchers: make(map[string]FetcherFunc),
		computes: make(map[string]ComputeFunc),
	}
}

// RegisterMeta registers a variable meta (overwrite allowed)
func (vc *VariableCenter) RegisterMeta(meta VariableMeta) {
	vc.metasMu.Lock()
	defer vc.metasMu.Unlock()
	vc.metas[meta.Key] = meta
}

// GetMeta gets meta, ok
func (vc *VariableCenter) GetMeta(key string) (VariableMeta, bool) {
	vc.metasMu.RLock()
	defer vc.metasMu.RUnlock()
	m, ok := vc.metas[key]
	return m, ok
}

// RegisterFetcher registers a named fetcher
func (vc *VariableCenter) RegisterFetcher(name string, fn FetcherFunc) {
	vc.fetchersMu.Lock()
	defer vc.fetchersMu.Unlock()
	vc.fetchers[name] = fn
}

// RegisterCompute registers a named compute function
func (vc *VariableCenter) RegisterCompute(name string, fn ComputeFunc) {
	vc.computesMu.Lock()
	defer vc.computesMu.Unlock()
	vc.computes[name] = fn
}

// ---------------------------
// Core: Get(key)
// ---------------------------

var ErrVariableNotFound = errors.New("variable meta not found")
var ErrCycleDetected = errors.New("cycle detected in variable dependencies")

// Get resolves a variable value for a VarContext.
// It handles cache, compute (depends), fetcher, TTL, and cycle detection.
func (vc *VariableCenter) Get(ctx context.Context, vctx *VarContext, key string) (any, error) {
	// 1) look meta
	meta, ok := vc.GetMeta(key)
	if !ok {
		// fallback: if key present in input, return it (useful convenience)
		if vctx != nil {
			if val, has := vctx.Input[key]; has {
				return val, nil
			}
		}
		return nil, ErrVariableNotFound
	}

	// 2) check per-request cache
	if meta.Cached && vctx != nil {
		if val, ok := vctx.getCached(key); ok {
			// append trace and return
			if vctx != nil {
				vctx.addTrace(key + " (cached)")
			}
			return val, nil
		}
	}

	// 3) cycle detection
	if vctx != nil {
		if vctx.visiting[key] {
			return nil, ErrCycleDetected
		}
		vctx.visiting[key] = true
		defer func() {
			delete(vctx.visiting, key)
		}()
	}

	// 4) If compute function exists, call it (after resolving dependencies if needed)
	if meta.ComputeName != "" {
		vc.computesMu.RLock()
		comp, ok := vc.computes[meta.ComputeName]
		vc.computesMu.RUnlock()
		if !ok {
			return nil, fmt.Errorf("compute function %s not registered", meta.ComputeName)
		}
		val, err := comp(ctx, vctx, meta, vc)
		if err != nil {
			return nil, err
		}
		// cache if allowed
		if meta.Cached && vctx != nil {
			vctx.setCache(key, val, meta.TTLSeconds)
		}
		if vctx != nil {
			vctx.addTrace(key)
		}
		return val, nil
	}

	// 5) else call fetcher
	if meta.FetcherName == "" {
		return nil, fmt.Errorf("no fetcher or compute for variable %s", key)
	}
	vc.fetchersMu.RLock()
	fetcher, ok := vc.fetchers[meta.FetcherName]
	vc.fetchersMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("fetcher %s not registered", meta.FetcherName)
	}
	val, err := fetcher(ctx, vctx, meta)
	if err != nil {
		return nil, err
	}
	if meta.Cached && vctx != nil {
		vctx.setCache(key, val, meta.TTLSeconds)
	}
	if vctx != nil {
		vctx.addTrace(key)
	}
	return val, nil
}

// ---------------------------
// Utilities: resolve dependencies helper
// ---------------------------

// ResolveDependencies fetches all variables in a list and returns a map (non-failing: returns error on first fail)
func (vc *VariableCenter) ResolveDependencies(ctx context.Context, vctx *VarContext, deps []string) (map[string]any, error) {
	out := make(map[string]any, len(deps))
	for _, k := range deps {
		val, err := vc.Get(ctx, vctx, k)
		if err != nil {
			return nil, err
		}
		out[k] = val
	}
	return out, nil
}

// ---------------------------
// Example: Built-in fetcher (input map)
// ---------------------------

func InputFetcher(ctx context.Context, vctx *VarContext, meta VariableMeta) (any, error) {
	if vctx == nil {
		return nil, fmt.Errorf("nil varcontext")
	}
	// direct key in input map usually stored under top-level key (e.g., "user.id")
	if val, ok := vctx.Input[meta.Key]; ok {
		return val, nil
	}
	// not found in input map
	return nil, fmt.Errorf("input key %s not present", meta.Key)
}

// ---------------------------
// Example: Built-in compute functions
// - compute via dependencies (we use a simple combiner demo)
// - in production you will register functions that run expr/CEL/goja etc.
// ---------------------------

// ExampleCompute_SumInts: assumes depends are "x" and "y" integer-like and returns sum
func ExampleCompute_SumInts(ctx context.Context, vctx *VarContext, meta VariableMeta, vcCenter *VariableCenter) (any, error) {
	// resolve deps
	deps, err := vcCenter.ResolveDependencies(ctx, vctx, meta.Depends)
	if err != nil {
		return nil, err
	}
	var sum int64 = 0
	for _, d := range deps {
		switch vv := d.(type) {
		case int:
			sum += int64(vv)
		case int64:
			sum += vv
		case float64:
			sum += int64(vv)
		default:
			// try to ignore if not numeric
		}
	}
	return sum, nil
}

// ExampleCompute_Concat: just concat string deps with sep
func ExampleCompute_Concat(ctx context.Context, vctx *VarContext, meta VariableMeta, vcCenter *VariableCenter) (any, error) {
	deps, err := vcCenter.ResolveDependencies(ctx, vctx, meta.Depends)
	if err != nil {
		return nil, err
	}
	out := ""
	for k, v := range deps {
		if out != "" {
			out += "|"
		}
		out += fmt.Sprintf("%v=%v", k, v)
	}
	return out, nil
}

// ---------------------------
// Demo: how to use inside a decision engine
// ---------------------------

func demoSequentialDecision(vc *VariableCenter) {
	fmt.Println("=== Demo: Sequential Decision using VariableCenter ===")

	// create per-request VarContext
	vctx := NewVarContext(map[string]any{
		"user.id":   "u-123",
		"user.name": "Alice",
		"x":         10,
		"y":         20,
	})

	// Decision logic (very simple): fetch device.riskScore and user.age then decide
	ctx := context.Background()

	// Assume variable keys "x_plus_y" is computed as sum of x and y
	val, err := vc.Get(ctx, vctx, "x_plus_y")
	if err != nil {
		fmt.Println("error get x_plus_y:", err)
		return
	}
	fmt.Println("x_plus_y =", val)

	// also fetch user.name (from input fetcher)
	name, err := vc.Get(ctx, vctx, "user.name")
	if err != nil {
		fmt.Println("error get user.name:", err)
		return
	}
	fmt.Println("user.name =", name)

	// show trace
	fmt.Println("Trace:", vctx.Trace)
}

func demoParallelAggregate(vc *VariableCenter) {
	fmt.Println("=== Demo: Parallel Aggregate using VariableCenter ===")
	vctx := NewVarContext(map[string]any{
		"x": 5,
		"y": 8,
	})

	ctx := context.Background()

	decisionKeys := []string{"x_plus_y", "x", "y"}
	var wg sync.WaitGroup
	results := make([]any, len(decisionKeys))
	errs := make([]error, len(decisionKeys))

	for i, k := range decisionKeys {
		wg.Add(1)
		go func(i int, key string) {
			defer wg.Done()
			vv, err := vc.Get(ctx, vctx, key)
			results[i] = vv
			errs[i] = err
		}(i, k)
	}
	wg.Wait()

	for i, k := range decisionKeys {
		fmt.Printf("key=%s val=%v err=%v\n", k, results[i], errs[i])
	}
	fmt.Println("Trace:", vctx.Trace)
}

// ---------------------------
// Main: register some metas and fetchers and run demo
// ---------------------------

func main() {
	vc := NewVariableCenter()

	// register input fetcher
	vc.RegisterFetcher("input", InputFetcher)

	// register compute functions (you'll replace these with expr/cel wrappers)
	vc.RegisterCompute("sumInts", ExampleCompute_SumInts)
	vc.RegisterCompute("concatDeps", ExampleCompute_Concat)

	// register metas
	vc.RegisterMeta(VariableMeta{
		Key:         "x",
		Name:        "x",
		Type:        TypeInt,
		FetcherName: "input",
		Source:      "request",
		Cached:      false,
	})
	vc.RegisterMeta(VariableMeta{
		Key:         "y",
		Name:        "y",
		Type:        TypeInt,
		FetcherName: "input",
		Source:      "request",
		Cached:      false,
	})
	vc.RegisterMeta(VariableMeta{
		Key:         "x_plus_y",
		Name:        "x+y",
		Type:        TypeInt,
		ComputeName: "sumInts",
		Depends:     []string{"x", "y"},
		Cached:      true,
		TTLSeconds:  30,
	})

	vc.RegisterMeta(VariableMeta{
		Key:         "user.name",
		Name:        "user.name",
		Type:        TypeString,
		FetcherName: "input",
		Cached:      false,
	})

	// run demos
	demoSequentialDecision(vc)
	fmt.Println()
	demoParallelAggregate(vc)
}
