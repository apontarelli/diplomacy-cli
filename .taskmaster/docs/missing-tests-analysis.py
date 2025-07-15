#!/usr/bin/env python3
"""
DATC Test Gap Analysis Script
Analyzes the DATC test JSON file and compares with implemented tests
"""

import json
import re
import os
import glob


def load_datc_tests():
    """Load DATC tests from JSON file"""
    datc_file = (
        "../backend/internal/game/pipeline/testdata/datc_tests_fixed.json"
    )
    with open(datc_file, "r") as f:
        data = json.load(f)
    return data["test_cases"]


def find_implemented_tests():
    """Find all implemented DATC tests in Go files"""
    implemented = set()
    test_files = glob.glob("../backend/internal/game/pipeline/*test.go")

    for file_path in test_files:
        with open(file_path, "r") as f:
            content = f.read()
            # Find test function names that match DATC pattern
            matches = re.findall(r"func Test.*DATC([A-Z])(\d+).*\(", content)
            for section, num in matches:
                test_id = f"6.{section}.{num}"
                implemented.add(test_id)

            # Also look for specific test IDs in comments or strings
            id_matches = re.findall(r"6\.[A-Z]\.\d+", content)
            for match in id_matches:
                implemented.add(match)

    return implemented


def categorize_tests(test_cases):
    """Categorize tests by section"""
    categories = {}
    for test in test_cases:
        test_id = test["id"]
        section = test["section"]
        category = test["category"]

        if section not in categories:
            categories[section] = {
                "tests": [],
                "category": category,
                "count": 0,
            }

        categories[section]["tests"].append(
            {
                "id": test_id,
                "name": test["name"],
                "description": test["description"],
            }
        )
        categories[section]["count"] += 1

    return categories


def analyze_gaps(test_cases, implemented_tests):
    """Analyze which tests are missing"""
    all_tests = {test["id"] for test in test_cases}
    missing_tests = all_tests - implemented_tests

    categories = categorize_tests(test_cases)

    print("=== DATC TEST GAP ANALYSIS ===\n")
    print(f"Total DATC Tests: {len(all_tests)}")
    print(f"Implemented Tests: {len(implemented_tests)}")
    print(f"Missing Tests: {len(missing_tests)}")
    print(
        f"Completion Rate: {len(implemented_tests) / len(all_tests) * 100:.1f}%\n"
    )

    print("=== CATEGORY BREAKDOWN ===\n")

    for section, data in sorted(categories.items()):
        section_tests = {test["id"] for test in data["tests"]}
        section_implemented = section_tests & implemented_tests
        section_missing = section_tests - implemented_tests

        status = (
            "✅ COMPLETE"
            if len(section_missing) == 0
            else f"🔴 {len(section_missing)} MISSING"
            if len(section_implemented) == 0
            else f"🟡 {len(section_missing)} MISSING"
        )

        print(f"**{section}** ({data['category']}) - {status}")
        print(
            f"  Total: {data['count']}, Implemented: {len(section_implemented)}, Missing: {len(section_missing)}"
        )

        if section_missing:
            print("  Missing Tests:")
            for test in data["tests"]:
                if test["id"] in section_missing:
                    print(f"    - {test['id']}: {test['name']}")
        print()

    print("=== PRIORITY RECOMMENDATIONS ===\n")

    # High priority categories (core game mechanics)
    high_priority = [
        "SUPPORTS AND DISLODGES",
        "HEAD-TO-HEAD BATTLES AND BELEAGUERED GARRISON",
    ]
    medium_priority = ["COASTAL ISSUES", "CONVOYS"]

    for section, data in sorted(categories.items()):
        section_tests = {test["id"] for test in data["tests"]}
        section_missing = section_tests - implemented_tests

        if section_missing:
            priority = (
                "HIGH"
                if section in high_priority
                else "MEDIUM"
                if section in medium_priority
                else "LOW"
            )
            print(f"**{section}** - Priority: {priority}")
            print(f"  {len(section_missing)} tests to implement")
            print(
                f"  Impact: {'Core game mechanics' if priority == 'HIGH' else 'Edge cases and accuracy' if priority == 'MEDIUM' else 'Nice to have'}"
            )
            print()


def main():
    # Change to script directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)

    test_cases = load_datc_tests()
    implemented_tests = find_implemented_tests()

    analyze_gaps(test_cases, implemented_tests)


if __name__ == "__main__":
    main()
