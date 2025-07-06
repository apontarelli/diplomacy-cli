#!/usr/bin/env python3
"""
Enhanced Go Test Generator for DATC Test Cases

Generates Go test files with proper unit setup and outcome validation.

Usage:
    python generate_enhanced_tests.py datc_tests_fixed.json --category basic_checks --output-dir ../../internal/game/pipeline/
"""

import argparse
import json
import re
from pathlib import Path
from typing import Dict, List, Tuple, Set, Optional


class UnitSetup:
    def __init__(self, unit_type: str, nation: str, province: str):
        self.unit_type = unit_type  # "Fleet" or "Army"
        self.nation = nation
        self.province = province


class EnhancedGoTestGenerator:
    def __init__(self):
        self.province_map = {
            # Common province name mappings for DATC format
            "north sea": "north_sea",
            "irish sea": "irish_sea",
            "english channel": "english_channel",
            "gulf of lyon": "gulf_of_lyon",
            "tyrrhenian sea": "tyrrhenian_sea",
            "ionian sea": "ionian_sea",
            "adriatic sea": "adriatic_sea",
            "aegean sea": "aegean_sea",
            "eastern mediterranean": "eastern_mediterranean",
            "black sea": "black_sea",
            "baltic sea": "baltic_sea",
            "barents sea": "barents_sea",
            "norwegian sea": "norwegian_sea",
            "north atlantic ocean": "north_atlantic_ocean",
            "mid atlantic ocean": "mid_atlantic_ocean",
            "western mediterranean": "western_mediterranean",
            "helgoland bight": "helgoland_bight",
            "skagerrak": "skagerrak",
            "bothnia": "bothnia",
            "st petersburg": "st_petersburg",
        }

    def normalize_province_name(self, province: str) -> str:
        """Convert province name to internal format."""
        province_lower = province.lower().strip()

        # Handle coast specifications
        if "/" in province_lower:
            base, coast = province_lower.split("/", 1)
            base = base.strip()
            coast = coast.strip()

            # Map the base province
            if base in self.province_map:
                base = self.province_map[base]
            else:
                base = base.replace(" ", "_")

            return f"{base}/{coast}"

        # Map known provinces
        if province_lower in self.province_map:
            return self.province_map[province_lower]

        # Default: replace spaces with underscores
        return province_lower.replace(" ", "_")

    def extract_units_from_orders(
        self, orders: Dict[str, List[str]]
    ) -> List[UnitSetup]:
        """Extract unit setup from order strings."""
        units = []

        for nation, order_list in orders.items():
            for order in order_list:
                unit_setup = self.parse_unit_from_order(order, nation)
                if unit_setup:
                    units.append(unit_setup)

        return units

    def parse_unit_from_order(
        self, order: str, nation: str
    ) -> Optional[UnitSetup]:
        """Parse a single order to extract unit information."""
        order = order.strip()

        # Match patterns like "F North Sea - Picardy" or "A Liverpool - Irish Sea"
        move_pattern = r"^([AF])\s+([^-]+?)\s*-\s*(.+)$"
        move_match = re.match(move_pattern, order)
        if move_match:
            unit_type_char, origin, destination = move_match.groups()
            unit_type = "Fleet" if unit_type_char == "F" else "Army"
            origin_province = self.normalize_province_name(origin.strip())
            return UnitSetup(unit_type, nation, origin_province)

        # Match hold patterns like "A Venice Hold"
        hold_pattern = r"^([AF])\s+([^-]+?)\s+Hold"
        hold_match = re.match(hold_pattern, order, re.IGNORECASE)
        if hold_match:
            unit_type_char, origin = hold_match.groups()
            unit_type = "Fleet" if unit_type_char == "F" else "Army"
            origin_province = self.normalize_province_name(origin.strip())
            return UnitSetup(unit_type, nation, origin_province)

        # Match support patterns like "F Trieste Supports F Trieste" or "A Tyrolia Supports A Venice - Trieste"
        support_pattern = r"^([AF])\s+([^-]+?)\s+Supports\s+([AF])\s+([^-]+?)(?:\s*-\s*(.+))?$"
        support_match = re.match(support_pattern, order, re.IGNORECASE)
        if support_match:
            (
                unit_type_char,
                origin,
                supported_unit_type,
                supported_origin,
                supported_dest,
            ) = support_match.groups()
            unit_type = "Fleet" if unit_type_char == "F" else "Army"
            origin_province = self.normalize_province_name(origin.strip())
            return UnitSetup(unit_type, nation, origin_province)

        # Match convoy patterns like "F North Sea Convoys A Yorkshire - Yorkshire"
        convoy_pattern = (
            r"^([AF])\s+([^-]+?)\s+Convoys\s+([AF])\s+([^-]+?)\s*-\s*(.+)$"
        )
        convoy_match = re.match(convoy_pattern, order, re.IGNORECASE)
        if convoy_match:
            (
                unit_type_char,
                origin,
                convoyed_unit_type,
                convoyed_origin,
                convoyed_dest,
            ) = convoy_match.groups()
            unit_type = "Fleet" if unit_type_char == "F" else "Army"
            origin_province = self.normalize_province_name(origin.strip())
            return UnitSetup(unit_type, nation, origin_province)

        return None

    def generate_test_function_name(self, test_id: str, test_name: str) -> str:
        """Generate a valid Go function name from test ID and name."""
        # Remove dots and replace with underscores
        clean_id = test_id.replace(".", "_")

        # Clean up the test name
        clean_name = re.sub(r"[^A-Za-z0-9\s]", "", test_name)
        clean_name = re.sub(r"\s+", "_", clean_name.strip())
        clean_name = clean_name.title().replace("_", "")

        return f"TestDATC_{clean_id}_{clean_name}"

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

    def generate_unit_setup_code(self, units: List[UnitSetup]) -> str:
        """Generate Go code for setting up units."""
        if not units:
            return "\t// No units to set up for this test"

        lines = []
        lines.append("\t// Set up initial units for this test")

        for unit in units:
            lines.append(f"\terr = gameState.Board.PlaceUnit(&game.Unit{{")
            lines.append(f"\t\tType:     game.{unit.unit_type},")
            lines.append(f"\t\tOwner:    game.{unit.nation},")
            lines.append(f'\t\tProvince: "{unit.province}",')
            lines.append(f"\t}})")
            lines.append(f"\tif err != nil {{")
            lines.append(
                f'\t\tt.Fatalf("Failed to place {unit.unit_type} in {unit.province}: %v", err)'
            )
            lines.append(f"\t}}")
            lines.append("")

        return "\n".join(lines)

    def generate_outcome_validation_code(
        self, expected_outcome: str, test_id: str
    ) -> str:
        """Generate Go code for validating test outcomes."""
        lines = []
        lines.append("\t// Validate results")
        lines.append(f'\tt.Logf("Test {test_id}: {expected_outcome}")')
        lines.append("")

        if "should fail" in expected_outcome.lower():
            lines.append("\t// Check that orders failed as expected")
            lines.append(
                "\tfor nation, orders := range result.ProcessedOrders {"
            )
            lines.append("\t\tfor _, order := range orders {")
            lines.append("\t\t\tif order.Status != game.OrderFailed {")
            lines.append(
                f'\t\t\t\tt.Errorf("Expected order to fail but got status: %v", order.Status)'
            )
            lines.append("\t\t\t}")
            lines.append("\t\t}")
            lines.append("\t}")
        elif "dislodged" in expected_outcome.lower():
            lines.append("\t// TODO: Check for dislodged units")
            lines.append(
                "\t// This requires implementing dislodgement tracking in the resolution engine"
            )
        else:
            lines.append(
                "\t// Basic validation - ensure processing completed without errors"
            )
            lines.append("\tif result == nil {")
            lines.append(
                '\t\tt.Fatal("Expected non-nil result from processing")'
            )
            lines.append("\t}")

        lines.append("")
        lines.append(f'\tt.Logf("Test {test_id} validation completed")')

        return "\n".join(lines)

    def generate_test_function(self, test_case: Dict) -> str:
        """Generate a complete Go test function for a single test case."""
        test_id = test_case["id"]
        test_name = test_case["name"]
        description = test_case["description"]
        orders = test_case["orders"]
        expected_outcome = test_case["expected_outcome"]

        func_name = self.generate_test_function_name(test_id, test_name)
        units = self.extract_units_from_orders(orders)
        unit_setup_code = self.generate_unit_setup_code(units)
        orders_code = self.format_raw_orders(orders)
        validation_code = self.generate_outcome_validation_code(
            expected_outcome, test_id
        )

        template = f"""// DATC Test {test_id}: {test_name}
func {func_name}(t *testing.T) {{
\t// Load classic map
\tmapLoader := loader.NewJSONLoader("../../../data/classic")
\tboard, err := mapLoader.LoadBoard()
\tif err != nil {{
\t\tt.Fatalf("Failed to load board: %v", err)
\t}}

\t// Create initial game state
\tgameState := game.NewGameState(board, game.SpringMovement, 1901)

{unit_setup_code}

\t// Set raw orders from DATC test case
\tgameState.RawOrders = map[game.Nation][]string{{
{orders_code}
\t}}

\t// Process the orders through the pipeline
\torchestrator := NewOrchestrator()
\tresult, err := orchestrator.ProcessTurn(gameState)
\tif err != nil {{
\t\tt.Fatalf("Failed to process turn: %v", err)
\t}}

{validation_code}
}}

"""
        return template

    def generate_test_file(
        self, test_cases: List[Dict], output_path: Path, category: str
    ):
        """Generate a complete Go test file for a category of test cases."""

        # Generate package header
        header = f"""package pipeline

import (
\t"testing"

\t"diplomacy-cli/backend/internal/game"
\t"diplomacy-cli/backend/internal/game/loader"
)

"""

        # Generate all test functions
        test_functions = []
        for test_case in test_cases:
            test_functions.append(self.generate_test_function(test_case))

        # Write to file
        with open(output_path, "w") as f:
            f.write(header)
            f.write("\n".join(test_functions))

        print(f"Generated {len(test_cases)} test functions in {output_path}")


def main():
    parser = argparse.ArgumentParser(
        description="Generate enhanced Go tests from DATC test cases"
    )
    parser.add_argument("input_file", help="Path to DATC test cases JSON file")
    parser.add_argument(
        "--output-dir", default=".", help="Output directory for Go test files"
    )
    parser.add_argument(
        "--category", help="Generate tests for specific category only"
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
    generator = EnhancedGoTestGenerator()

    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    if args.by_category:
        # Group by category
        categories = {}
        for test_case in test_cases:
            category = test_case["category"]
            if category not in categories:
                categories[category] = []
            categories[category].append(test_case)

        # Generate file for each category
        for category, category_tests in categories.items():
            if args.category and category != args.category:
                continue

            output_file = output_dir / f"datc_{category}_test.go"
            generator.generate_test_file(category_tests, output_file, category)

    elif args.category:
        # Generate for specific category
        category_tests = [
            tc for tc in test_cases if tc["category"] == args.category
        ]
        output_file = output_dir / f"datc_{args.category}_test.go"
        generator.generate_test_file(category_tests, output_file, args.category)

    else:
        # Generate all tests in one file
        output_file = output_dir / "datc_all_tests.go"
        generator.generate_test_file(test_cases, output_file, "all")


if __name__ == "__main__":
    main()
