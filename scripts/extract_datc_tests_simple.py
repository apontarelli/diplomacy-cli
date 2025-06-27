#!/usr/bin/env python3
"""
Simple DATC test case extractor that generates basic test stubs.
"""

import re
from pathlib import Path

def generate_simple_datc_tests():
    """Generate simple DATC test stubs for each category"""
    
    categories = {
        'A': ('BasicChecks', 12),
        'B': ('CoastalIssues', 15), 
        'C': ('CircularMovement', 9),
        'D': ('SupportsAndDislodges', 34),
        'E': ('HeadToHeadBattles', 14),
        'F': ('Convoys', 25),
        'G': ('ConvoyingToAdjacent', 20),
        'H': ('Retreating', 16),
        'I': ('Building', 7),
        'J': ('CivilDisorder', 11)
    }
    
    output_dir = Path("backend/internal/api")
    output_dir.mkdir(parents=True, exist_ok=True)
    
    for category, (name, count) in categories.items():
        content = f'''package api

import (
\t"testing"
)

// DATC Test Cases - Category {category}: {name}
// Generated stubs for {count} test cases from DATC v3.0
// TODO: Implement actual test cases from https://webdiplomacy.net/doc/DATC_v3_0.html

'''
        
        for i in range(1, count + 1):
            func_name = f"TestDATC_6_{category}_{i}"
            content += f'''func {func_name}(t *testing.T) {{
\tt.Skip("DATC test case 6.{category}.{i} not yet implemented")
\t
\t// TODO: Implement DATC test case 6.{category}.{i}
\t// See: https://webdiplomacy.net/doc/DATC_v3_0.html#6.{category}.{i}
}}

'''
        
        filename = f"datc_{category.lower()}_test.go"
        filepath = output_dir / filename
        
        with open(filepath, 'w') as f:
            f.write(content)
        
        print(f"Generated {filepath} with {count} test stubs")
    
    print(f"\nGenerated {len(categories)} DATC test files with {sum(count for _, count in categories.values())} test stubs")
    print("Next steps:")
    print("1. Run tests to verify they compile: go test ./backend/internal/api/...")
    print("2. Implement test cases one by one using the DATC documentation")
    print("3. Remove t.Skip() calls as tests are implemented")

if __name__ == "__main__":
    generate_simple_datc_tests()