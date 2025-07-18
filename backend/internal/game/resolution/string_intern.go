package resolution

import (
	"fmt"
	"strings"
	"sync"
)

// StringInterner provides string interning to reduce memory allocations
// for frequently used strings like territory names, order types, and cache keys
type StringInterner struct {
	strings   map[string]string
	accessLog []string // Simple LRU tracking
	maxSize   int
	mu        sync.RWMutex
}

const (
	// DefaultMaxInternedStrings is the default maximum number of strings to intern
	DefaultMaxInternedStrings = 10000
)

// NewStringInterner creates a new string interner with default size limit
func NewStringInterner() *StringInterner {
	return NewStringInternerWithLimit(DefaultMaxInternedStrings)
}

// NewStringInternerWithLimit creates a new string interner with specified size limit
func NewStringInternerWithLimit(maxSize int) *StringInterner {
	return &StringInterner{
		strings:   make(map[string]string),
		accessLog: make([]string, 0, maxSize/10), // Small initial capacity for access log
		maxSize:   maxSize,
	}
}

// Intern returns the canonical instance of the string, reducing memory usage
func (si *StringInterner) Intern(s string) string {
	if s == "" {
		return ""
	}

	si.mu.RLock()
	if interned, exists := si.strings[s]; exists {
		si.mu.RUnlock()
		return interned
	}
	si.mu.RUnlock()

	si.mu.Lock()
	defer si.mu.Unlock()

	// Double-check after acquiring write lock
	if interned, exists := si.strings[s]; exists {
		si.updateAccessLog(s)
		return interned
	}

	// Check if we need to evict old entries
	if len(si.strings) >= si.maxSize {
		si.evictLRU()
	}

	// Store the string and return it
	si.strings[s] = s
	si.updateAccessLog(s)
	return s
}

// Size returns the number of interned strings
func (si *StringInterner) Size() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.strings)
}

// Clear removes all interned strings
func (si *StringInterner) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.strings = make(map[string]string)
	si.accessLog = si.accessLog[:0] // Clear access log but keep capacity
}

// updateAccessLog updates the access log for LRU tracking (must be called with write lock held)
func (si *StringInterner) updateAccessLog(s string) {
	if s == "" {
		return // Don't track empty strings
	}

	// Simple LRU: move to end of access log
	// Remove existing entry if present
	for i, entry := range si.accessLog {
		if entry == s {
			// Move to end by removing and appending
			si.accessLog = append(si.accessLog[:i], si.accessLog[i+1:]...)
			break
		}
	}

	// Prevent access log from growing too large
	if len(si.accessLog) >= si.maxSize {
		si.accessLog = si.accessLog[1:] // Remove oldest entry
	}

	si.accessLog = append(si.accessLog, s)
}

// evictLRU removes the least recently used entries (must be called with write lock held)
func (si *StringInterner) evictLRU() {
	if len(si.accessLog) == 0 {
		return
	}

	// Evict 10% of entries to avoid frequent evictions
	evictCount := si.maxSize / 10
	if evictCount < 1 {
		evictCount = 1
	}

	for i := 0; i < evictCount && len(si.accessLog) > 0; i++ {
		// Remove least recently used (first in access log)
		lru := si.accessLog[0]
		delete(si.strings, lru)
		si.accessLog = si.accessLog[1:]
	}
}

// Global string interner for common values
var globalInterner = NewStringInterner()

// Common string constants that are frequently used
const (
	// Order type strings (pre-interned)
	MoveString    = "Move"
	SupportString = "Support"
	ConvoyString  = "Convoy"
	HoldString    = "Hold"
	RetreatString = "Retreat"
	UnknownString = "Unknown"

	// Common separators and formats
	ArrowSeparator    = " -> "
	DashSeparator     = " - "
	OptimisticSuffix  = ":opt"
	PessimisticSuffix = ":pes"

	// Common reason prefixes
	OrderSucceededPrefix  = " order succeeded"
	OrderFailedPrefix     = " order failed"
	MoveSucceededPrefix   = "Move succeeded"
	MoveFailedPrefix      = "Move failed"
	SupportCutPrefix      = "Support cut by successful attack from "
	ConvoyDisruptedPrefix = "Convoy disrupted by successful attack from "
)

// Pre-intern common strings at package initialization
func init() {
	// Pre-intern order type strings
	globalInterner.Intern(MoveString)
	globalInterner.Intern(SupportString)
	globalInterner.Intern(ConvoyString)
	globalInterner.Intern(HoldString)
	globalInterner.Intern(RetreatString)
	globalInterner.Intern(UnknownString)

	// Pre-intern common separators
	globalInterner.Intern(ArrowSeparator)
	globalInterner.Intern(DashSeparator)
	globalInterner.Intern(OptimisticSuffix)
	globalInterner.Intern(PessimisticSuffix)

	// Pre-intern common reason prefixes
	globalInterner.Intern(OrderSucceededPrefix)
	globalInterner.Intern(OrderFailedPrefix)
	globalInterner.Intern(MoveSucceededPrefix)
	globalInterner.Intern(MoveFailedPrefix)
	globalInterner.Intern(SupportCutPrefix)
	globalInterner.Intern(ConvoyDisruptedPrefix)
}

// InternString interns a string using the global interner
func InternString(s string) string {
	return globalInterner.Intern(s)
}

// InternTerritory interns a territory name
func InternTerritory(territory string) string {
	return globalInterner.Intern(territory)
}

// InternUnit interns a unit string
func InternUnit(unit string) string {
	return globalInterner.Intern(unit)
}

// InternOwner interns an owner/nation string
func InternOwner(owner string) string {
	return globalInterner.Intern(owner)
}

// StringBuilderPool provides a pool of string builders to reduce allocations
type StringBuilderPool struct {
	pool sync.Pool
}

// NewStringBuilderPool creates a new string builder pool
func NewStringBuilderPool() *StringBuilderPool {
	return &StringBuilderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
	}
}

// Get retrieves a string builder from the pool
func (sbp *StringBuilderPool) Get() *strings.Builder {
	return sbp.pool.Get().(*strings.Builder)
}

// Put returns a string builder to the pool after resetting it
func (sbp *StringBuilderPool) Put(sb *strings.Builder) {
	sb.Reset()
	sbp.pool.Put(sb)
}

// Global string builder pool
var globalBuilderPool = NewStringBuilderPool()

// GetStringBuilder gets a string builder from the global pool
func GetStringBuilder() *strings.Builder {
	return globalBuilderPool.Get()
}

// PutStringBuilder returns a string builder to the global pool
func PutStringBuilder(sb *strings.Builder) {
	globalBuilderPool.Put(sb)
}

// BuildCacheKey efficiently builds cache keys using string builder and interning with validation
func BuildCacheKey(parts ...string) string {
	if len(parts) == 0 {
		return InternString("empty-key")
	}

	// Validate all parts are non-empty
	for _, part := range parts {
		if part == "" {
			return InternString("invalid-key")
		}
	}

	if len(parts) == 1 {
		return InternString(parts[0])
	}

	sb := GetStringBuilder()
	defer PutStringBuilder(sb)

	for i, part := range parts {
		if i > 0 {
			sb.WriteString(":")
		}
		sb.WriteString(part)
	}

	return InternString(sb.String())
}

// BuildArrowKey builds a "source -> destination" key efficiently with validation
func BuildArrowKey(source, destination string) string {
	// Validate inputs to prevent malformed keys
	if source == "" || destination == "" {
		return InternString("invalid-key")
	}

	sb := GetStringBuilder()
	defer PutStringBuilder(sb)

	sb.WriteString(source)
	sb.WriteString(ArrowSeparator)
	sb.WriteString(destination)

	return InternString(sb.String())
}

// BuildOptimisticKey builds a cache key with optimistic suffix
func BuildOptimisticKey(base string, optimistic bool) string {
	sb := GetStringBuilder()
	defer PutStringBuilder(sb)

	sb.WriteString(base)
	if optimistic {
		sb.WriteString(OptimisticSuffix)
	} else {
		sb.WriteString(PessimisticSuffix)
	}

	return InternString(sb.String())
}

// SafeStringBuilder provides error-safe string building operations
type SafeStringBuilder struct {
	sb  *strings.Builder
	err error
}

// NewSafeStringBuilder creates a new error-safe string builder
func NewSafeStringBuilder() *SafeStringBuilder {
	return &SafeStringBuilder{
		sb: GetStringBuilder(),
	}
}

// WriteString safely writes a string, capturing any errors
func (ssb *SafeStringBuilder) WriteString(s string) *SafeStringBuilder {
	if ssb.err != nil {
		return ssb // Already in error state
	}

	_, ssb.err = ssb.sb.WriteString(s)
	return ssb
}

// WriteInt safely writes an integer as string
func (ssb *SafeStringBuilder) WriteInt(i int) *SafeStringBuilder {
	if ssb.err != nil {
		return ssb
	}

	_, ssb.err = ssb.sb.WriteString(fmt.Sprintf("%d", i))
	return ssb
}

// String returns the built string and any error, cleaning up the builder
func (ssb *SafeStringBuilder) String() (string, error) {
	defer PutStringBuilder(ssb.sb)

	if ssb.err != nil {
		return "", ssb.err
	}

	return InternString(ssb.sb.String()), nil
}

// MustString returns the built string, panicking on error (for cases where errors are unexpected)
func (ssb *SafeStringBuilder) MustString() string {
	result, err := ssb.String()
	if err != nil {
		panic(fmt.Sprintf("unexpected string building error: %v", err))
	}
	return result
}
