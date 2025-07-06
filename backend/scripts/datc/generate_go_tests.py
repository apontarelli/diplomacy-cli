#!/usr/bin/env python3
"""
Generate Go test files from DATC test cases.

This script creates Go test files that can be used to validate our Diplomacy engine
against the official DATC test cases.
"""

import argparse
import json
import os
import re
from pathlib import Path
from typing import Dict, List


class GoTestGenerator:
    def __init__(self):
        self.test_template = """package datc

import (
	"testing"
	
	"github.com/diplomacy-cli/backend/internal/game"
	"github.com/diplomacy-cli/backend/internal/game/loader"
	"github.com/diplomacy-cli/backend/internal/pipeline"
)

// {test_description}
func Test{test_function_name}(t *testing.T) {{
	// Load classic map
	mapLoader := loader.NewJSONLoader()
	board, err := mapLoader.LoadMap("../../../data/classic")
	if err != nil {{
		t.Fatalf("Failed to load map: %v", err)
	}}

	// Create initial game state
	gameState := game.NewGameState(board)
	
	// Set up initial units (this would need to be customized per test)
	// TODO: Parse initial setup from test case
	
	// Raw orders from DATC test case
	rawOrders := map[game.Nation][]string{{
{raw_orders}
	}}
	
	// Process orders through pipeline
	orchestrator := pipeline.NewOrchestrator()
	result, err := orchestrator.ProcessTurn(gameState, rawOrders)
	if err != nil {{
		t.Fatalf("Failed to process turn: %v", err)
	}}
	
	// Validate results
	// TODO: Implement specific validation for this test case
	// Expected: {expected_outcome}
	
	_ = result // Prevent unused variable error for now
	
	// This is a placeholder - actual validation logic would go here
	t.Logf("Test {test_id} completed - manual validation required")
	t.Logf("Expected: {expected_outcome}")
}}

"""

    def generate_test_function_name(self, test_id: str, test_name: str) -> str:
        """Generate a valid Go function name from test ID and name."""
        # Remove dots and replace with underscores
        clean_id = test_id.replace(".", "_")

        # Clean up the test name
        clean_name = re.sub(r"[^A-Za-z0-9\s]", "", test_name)
        clean_name = re.sub(r"\s+", "_", clean_name.strip())
        clean_name = clean_name.title().replace("_", "")

        return f"DATC_{clean_id}_{clean_name}"

    def format_raw_orders(self, orders: Dict[str, List[str]]) -> str:
        """Format orders for Go map literal."""
        if not orders:
            return ""

        lines = []
        for nation, order_list in orders.items():
            if order_list:
                formatted_orders = ", ".join(
                    f'"{order}"' for order in order_list
                )
                lines.append(f"\t\tgame.{nation}: {{{formatted_orders}}},")

        return "\n".join(lines)

    def generate_test_file(
        self, test_cases: List[Dict], output_path: Path
    ) -> None:
        """Generate a Go test file for a list of test cases."""

        if not test_cases:
            return

        # Generate file header
        file_header = """package datc

import (
	"testing"
	
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
	"diplomacy-cli/backend/internal/pipeline"
)

"""

        # Generate individual test functions (without package/imports)
        test_function_template = """// {test_description}
func Test{test_function_name}(t *testing.T) {{
	// Load classic map
	mapLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {{
		t.Fatalf("Failed to load board: %v", err)
	}}

	// Create initial game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)
	
	// Set up initial units (this would need to be customized per test)
	// TODO: Parse initial setup from test case and add units to gameState
	
	// Set raw orders from DATC test case
	gameState.RawOrders = map[game.Nation][]string{{
{raw_orders}
	}}
	
	// Process orders through pipeline
	processor := pipeline.NewTurnProcessor()
	result, err := processor.ProcessTurn(gameState)
	if err != nil {{
		t.Fatalf("Failed to process turn: %v", err)
	}}
	
	// Validate results
	// TODO: Implement specific validation for this test case
	// Expected: {expected_outcome}
	
	_ = result // Prevent unused variable error for now
	
	// This is a placeholder - actual validation logic would go here
	t.Logf("Test {test_id} completed - manual validation required")
	t.Logf("Expected: {expected_outcome}")
}}

"""

        # Generate individual test functions (without package/imports)
        test_function_template = """// {test_description}
func Test{test_function_name}(t *testing.T) {{
	// Load classic map
	mapLoader := loader.NewJSONLoader()
	board, err := mapLoader.LoadMap("../../../data/classic")
	if err != nil {{
		t.Fatalf("Failed to load map: %v", err)
	}}

	// Create initial game state
	gameState := game.NewGameState(board)
	
	// Set up initial units (this would need to be customized per test)
	// TODO: Parse initial setup from test case
	
	// Raw orders from DATC test case
	rawOrders := map[game.Nation][]string{{
{raw_orders}
	}}
	
	// Process orders through pipeline
	orchestrator := pipeline.NewOrchestrator()
	result, err := orchestrator.ProcessTurn(gameState, rawOrders)
	if err != nil {{
		t.Fatalf("Failed to process turn: %v", err)
	}}
	
	// Validate results
	// TODO: Implement specific validation for this test case
	// Expected: {expected_outcome}
	
	_ = result // Prevent unused variable error for now
	
	// This is a placeholder - actual validation logic would go here
	t.Logf("Test {test_id} completed - manual validation required")
	t.Logf("Expected: {expected_outcome}")
}}

"""

        test_functions = []

        for test_case in test_cases:
            test_id = test_case["id"]
            test_name = test_case["name"]
            description = test_case.get("description", "No description")
            orders = test_case.get("orders", {})
            expected_outcome = test_case.get(
                "expected_outcome", "No expected outcome specified"
            )

            function_name = self.generate_test_function_name(test_id, test_name)
            raw_orders = self.format_raw_orders(orders)

            test_function = test_function_template.format(
                test_description=f"DATC Test {test_id}: {test_name}",
                test_function_name=function_name,
                test_id=test_id,
                raw_orders=raw_orders,
                expected_outcome=expected_outcome,
            )

            test_functions.append(test_function)

        # Write to file
        with open(output_path, "w") as f:
            f.write(file_header)
            f.write("\n".join(test_functions))

        print(f"Generated {len(test_cases)} test functions in {output_path}")

    def generate_by_category(
        self, test_cases: List[Dict], output_dir: Path
    ) -> None:
        """Generate separate test files for each category."""

        # Group by category
        by_category = {}
        for test_case in test_cases:
            category = test_case.get("category", "unknown")
            if category not in by_category:
                by_category[category] = []
            by_category[category].append(test_case)

        # Generate file for each category
        for category, cases in by_category.items():
            filename = f"datc_{category}_test.go"
            output_path = output_dir / filename
            self.generate_test_file(cases, output_path)

        print(f"\nGenerated {len(by_category)} category files:")
        for category, cases in by_category.items():
            print(f"  {category}: {len(cases)} tests")

    def generate_subset(
        self,
        test_cases: List[Dict],
        start_idx: int,
        count: int,
        output_path: Path,
    ) -> None:
        """Generate tests for a subset of test cases."""
        subset = test_cases[start_idx : start_idx + count]
        self.generate_test_file(subset, output_path)


def main():
    parser = argparse.ArgumentParser(
        description="Generate Go tests from DATC test cases"
    )
    parser.add_argument(
        "input_file", help="JSON file with extracted DATC test cases"
    )
    parser.add_argument(
        "--output-dir",
        "-o",
        default=".",
        help="Output directory for Go test files",
    )
    parser.add_argument(
        "--category", help="Generate tests for specific category only"
    )
    parser.add_argument(
        "--subset", help="Generate subset: start,count (e.g., 0,5)"
    )
    parser.add_argument(
        "--by-category",
        action="store_true",
        help="Generate separate files by category",
    )

    args = parser.parse_args()

    # Load test cases
    with open(args.input_file, "r") as f:
        data = json.load(f)

    test_cases = data["test_cases"]
    output_dir = Path(args.output_dir)
    output_dir.mkdir(exist_ok=True)

    generator = GoTestGenerator()

    if args.subset:
        start, count = map(int, args.subset.split(","))
        output_path = output_dir / f"datc_subset_{start}_{count}_test.go"
        generator.generate_subset(test_cases, start, count, output_path)
    elif args.category:
        filtered_cases = [
            tc for tc in test_cases if tc.get("category") == args.category
        ]
        output_path = output_dir / f"datc_{args.category}_test.go"
        generator.generate_test_file(filtered_cases, output_path)
    elif args.by_category:
        generator.generate_by_category(test_cases, output_dir)
    else:
        # Generate all tests in one file
        output_path = output_dir / "datc_all_test.go"
        generator.generate_test_file(test_cases, output_path)


if __name__ == "__main__":
    main()
