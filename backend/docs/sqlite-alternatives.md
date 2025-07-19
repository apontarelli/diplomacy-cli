# SQLite Driver Alternatives

## Current: github.com/mattn/go-sqlite3

**What it is:** CGO-based SQLite driver (wraps C SQLite library)

**Pros:**
- ✅ Full SQLite feature support
- ✅ Best performance (native C code)
- ✅ Most mature and widely used
- ✅ Supports all SQLite extensions

**Cons:**
- ❌ Requires CGO (C compiler needed)
- ❌ Cross-compilation complexity
- ❌ Larger binary size
- ❌ Platform-specific builds

## Alternative 1: modernc.org/sqlite (Pure Go)

**What it is:** Pure Go SQLite implementation (no CGO)

```go
import (
    "database/sql"
    _ "modernc.org/sqlite"
)

// Same API, just change the driver name
db, err := sql.Open("sqlite", "diplomacy.db")
```

**Pros:**
- ✅ Pure Go (no CGO required)
- ✅ Easy cross-compilation
- ✅ Same database/sql interface
- ✅ Smaller deployment footprint

**Cons:**
- ❌ ~20% slower than CGO version
- ❌ Less mature (newer project)
- ❌ Some SQLite extensions not supported

## Alternative 2: Embedded Database Alternatives

### Option A: BadgerDB (Pure Go Key-Value)
```go
// Would require rewriting all SQL to key-value operations
db, err := badger.Open(badger.DefaultOptions("./data"))
```

### Option B: BoltDB/bbolt (Pure Go Key-Value)
```go
// Would require complete data model rewrite
db, err := bolt.Open("diplomacy.db", 0600, nil)
```

### Option C: In-Memory with Persistence
```go
// Use maps + JSON/gob serialization
type GameStore struct {
    games   map[int64]*Game
    players map[int64]*Player
    // ... serialize to disk periodically
}
```

## Recommendation Analysis

### For Diplomacy CLI Project:

**Keep github.com/mattn/go-sqlite3** because:

1. **SQL Compatibility**: We're already using complex SQL with:
   - Foreign keys
   - Triggers
   - Indexes
   - Transactions
   - JOINs

2. **Migration Cost**: Switching would require:
   - Rewriting all SQL queries
   - Redesigning data access patterns
   - Extensive testing
   - No functional benefit

3. **Performance**: Game state queries benefit from SQL optimization

4. **Deployment**: CGO is acceptable for a game server (not a CLI tool)

### If CGO is Absolutely Unacceptable:

**Switch to modernc.org/sqlite:**
```bash
go mod edit -replace github.com/mattn/go-sqlite3=modernc.org/sqlite@latest
# Change driver name in database.go from "sqlite3" to "sqlite"
```

**Minimal code changes required** - just the driver import and connection string.

## Current Usage Analysis

Our current SQL usage:
- Complex schema with 7 tables
- Foreign key constraints
- Triggers for timestamps
- Performance indexes
- Transaction support

**Verdict**: SQLite driver is justified for this use case. The alternative would be a complete architectural rewrite for minimal benefit.