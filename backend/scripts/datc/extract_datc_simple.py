#!/usr/bin/env python3
"""
Simple DATC Test Case Extractor

Extracts test cases from DATC v3.0 HTML using only standard library.
This is a simplified version that focuses on getting the basic structure.

Usage:
    python extract_datc_simple.py datc_v3_0.html
"""

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Dict, List, Optional


class SimpleDATCExtractor:
    def __init__(self):
        self.test_cases = []

    def extract_from_html(self, html_path: Path) -> List[Dict]:
        """Extract test cases from HTML file using regex patterns."""
        print(f"Reading HTML from {html_path}...")

        try:
            with open(html_path, "r", encoding="utf-8") as f:
                content = f.read()
        except UnicodeDecodeError:
            with open(html_path, "r", encoding="latin-1") as f:
                content = f.read()

        print(f"HTML file size: {len(content):,} characters")

        # Extract test cases directly from HTML (don't strip tags first)
        test_cases = self._extract_test_cases_from_html(content)

        print(f"Extracted {len(test_cases)} test cases")
        return test_cases

    def _strip_html_tags(self, html: str) -> str:
        """Remove HTML tags and clean up text."""
        # Remove HTML tags
        text = re.sub(r"<[^>]+>", "", html)

        # Decode HTML entities
        text = text.replace("&nbsp;", " ")
        text = text.replace("&amp;", "&")
        text = text.replace("&lt;", "<")
        text = text.replace("&gt;", ">")
        text = text.replace("&quot;", '"')

        # Clean up whitespace
        text = re.sub(r"\s+", " ", text)
        text = re.sub(r"\n\s*\n", "\n\n", text)

        return text

    def _extract_test_cases_from_html(self, html: str) -> List[Dict]:
        """Extract test cases using regex patterns."""
        test_cases = []

        # Find all individual test cases directly (6.A.1, 6.A.2, etc.)
        test_id_pattern = r'<h4><a name="(6\.[A-J]\.\d+)">'
        test_ids = re.findall(test_id_pattern, html)

        print(f"Found {len(test_ids)} individual test cases")

        for test_id in test_ids:
            # Extract the full header for this test case
            header_pattern = (
                rf'<h4><a name="{re.escape(test_id)}">(.*?)</a></h4>'
            )
            header_match = re.search(header_pattern, html, re.DOTALL)

            if not header_match:
                continue

            test_header = header_match.group(1).strip()

            # Extract the test name from the header
            test_name_match = re.search(r"TEST CASE[,\s]*(.+)", test_header)
            test_name = (
                test_name_match.group(1).strip()
                if test_name_match
                else "Unknown"
            )

            # Determine section and category
            section_letter = test_id.split(".")[1]
            section_id = f"6.{section_letter}"

            print(f"Processing test case {test_id}: {test_name[:50]}...")

            # Find the content for this specific test case
            test_content = self._extract_test_case_content_by_id(html, test_id)

            if test_content:
                test_case = self._parse_test_case(
                    test_id,
                    test_name,
                    self._get_section_title(section_letter),
                    test_content,
                )
                if test_case:
                    test_cases.append(test_case)

        return test_cases

    def _extract_test_case_content_by_id(
        self, text: str, test_id: str
    ) -> Optional[str]:
        """Extract content for a specific test case by ID."""
        # Find the start of this test case
        test_start_pattern = rf'<h4><a name="{re.escape(test_id)}">'
        test_start_match = re.search(test_start_pattern, text)

        if not test_start_match:
            print(f"    Warning: Could not find test case {test_id}")
            return None

        start_pos = test_start_match.start()

        # Find the next test case or end of section
        next_test_pattern = r'<h4><a name="6\.[A-J]\.\d+">'
        next_test_match = re.search(next_test_pattern, text[start_pos + 1 :])

        if next_test_match:
            end_pos = start_pos + 1 + next_test_match.start()
        else:
            end_pos = len(text)

        content = text[start_pos:end_pos]
        print(f"    Extracted {len(content)} characters for {test_id}")
        return content

    def _get_section_title(self, section_letter: str) -> str:
        """Get section title from section letter."""
        section_titles = {
            "A": "BASIC CHECKS",
            "B": "COASTAL ISSUES",
            "C": "CIRCULAR MOVEMENT",
            "D": "SUPPORTS AND DISLODGES",
            "E": "HEAD-TO-HEAD BATTLES AND BELEAGUERED GARRISON",
            "F": "CONVOYS",
            "G": "CONVOYING TO ADJACENT PROVINCES",
            "H": "RETREATING",
            "I": "BUILDING",
            "J": "CIVIL DISORDER AND DISBANDS",
        }
        return section_titles.get(section_letter, f"SECTION {section_letter}")

    def _extract_section_test_cases(
        self, section_id: str, section_title: str, content: str
    ) -> List[Dict]:
        """Extract individual test cases from a section."""
        test_cases = []

        # Pattern for individual test cases like "6.A.1. TEST CASE, MOVING TO..."
        test_pattern = (
            rf"{re.escape(section_id)}\.(\d+)\.\\s+TEST CASE[,\\s]*([^\\n]*)"
        )
        test_matches = re.findall(test_pattern, content, re.IGNORECASE)

        print(
            f"  Found {len(test_matches)} test case matches in section {section_id}"
        )

        for test_number, test_name in test_matches:
            test_id = f"{section_id}.{test_number}"

            # Find the content for this specific test case
            test_content = self._extract_test_case_content(
                content, test_id, test_name
            )

            if test_content:
                test_case = self._parse_test_case(
                    test_id, test_name, section_title, test_content
                )
                if test_case:
                    test_cases.append(test_case)
                    print(
                        f"    Parsed test case {test_id}: {test_name[:50]}..."
                    )

        return test_cases

    def _extract_test_case_content(
        self, section_content: str, test_id: str, test_name: str
    ) -> Optional[str]:
        """Extract content for a specific test case."""
        # Find the start of this test case
        test_header = f"{test_id}. TEST CASE"
        start_pos = section_content.find(test_header)

        if start_pos == -1:
            print(
                f"    Warning: Could not find header '{test_header}' in section"
            )
            return None

        # Find the end (next test case or end of section)
        # Look for the next test case in the same section or any section
        next_test_pattern = rf"6\.[A-J]\.\d+\.\s+TEST CASE"
        search_text = section_content[start_pos + len(test_header) :]
        next_test_match = re.search(
            next_test_pattern, search_text, re.IGNORECASE
        )

        if next_test_match:
            end_pos = start_pos + len(test_header) + next_test_match.start()
        else:
            end_pos = len(section_content)

        content = section_content[start_pos:end_pos].strip()
        print(f"    Extracted {len(content)} characters for {test_id}")
        return content

    def _parse_test_case(
        self, test_id: str, test_name: str, section_title: str, content: str
    ) -> Optional[Dict]:
        """Parse a single test case content."""

        # Extract description from <p> tags
        description = self._extract_description_from_html(content)

        # Extract orders from <pre> tags
        orders = self._extract_orders_from_html(content)

        # Extract expected outcome from <p> tags
        expected_outcome = self._extract_expected_outcome_from_html(content)

        # Determine category
        category = self._get_category_from_section(section_title)

        test_case = {
            "id": test_id,
            "name": test_name.strip(),
            "section": section_title.strip(),
            "category": category,
            "description": description,
            "orders": orders,
            "expected_outcome": expected_outcome,
            "raw_content": content[:500] + "..."
            if len(content) > 500
            else content,
        }

        return test_case

    def _extract_description_from_html(self, content: str) -> str:
        """Extract description from HTML <p> tags."""
        # Find the first <p> tag after the header
        p_pattern = r"<p>([^<]+)</p>"
        p_matches = re.findall(p_pattern, content, re.DOTALL)

        if p_matches:
            # Clean up the text
            description = p_matches[0].strip()
            description = re.sub(r"\s+", " ", description)
            return description

        return "No description found"

    def _extract_orders_from_html(self, content: str) -> Dict[str, List[str]]:
        """Extract orders from HTML <pre> tags."""
        orders = {}

        # Find all <pre> blocks
        pre_pattern = r"<pre>([^<]+)</pre>"
        pre_matches = re.findall(pre_pattern, content, re.DOTALL)

        for pre_content in pre_matches:
            lines = pre_content.strip().split("\n")
            current_nation = None

            for line in lines:
                line = line.strip()
                if not line:
                    continue

                # Check for nation header (e.g., "England:")
                nation_match = re.match(
                    r"^(England|France|Germany|Austria|Italy|Russia|Turkey):\s*$",
                    line,
                )
                if nation_match:
                    current_nation = nation_match.group(1)
                    if current_nation not in orders:
                        orders[current_nation] = []
                    continue

                # If we have a current nation and this looks like an order
                if current_nation and line:
                    orders[current_nation].append(line)

        return orders

    def _extract_expected_outcome_from_html(self, content: str) -> str:
        """Extract expected outcome from HTML <p> tags."""
        # Find all <p> tags
        p_pattern = r"<p>([^<]+)</p>"
        p_matches = re.findall(p_pattern, content, re.DOTALL)

        outcome_keywords = [
            "should fail",
            "fails",
            "succeeds",
            "dislodged",
            "bounces",
            "will move",
            "will not move",
            "advances",
            "order should",
        ]

        for p_content in p_matches:
            p_text = p_content.strip()
            if any(keyword in p_text.lower() for keyword in outcome_keywords):
                # Clean up the text
                outcome = re.sub(r"\s+", " ", p_text)
                return outcome

        return "No outcome specified"

    def _extract_description(self, content: str) -> str:
        """Extract the test case description."""
        lines = content.split("\n")

        # Skip the header line and find the first substantial description
        for line in lines[1:]:
            line = line.strip()
            if line and len(line) > 20:
                # Skip lines that look like orders or outcomes
                if not re.match(
                    r"^(England|France|Germany|Austria|Italy|Russia|Turkey):",
                    line,
                ):
                    if not any(
                        keyword in line.lower()
                        for keyword in [
                            "order should",
                            "move fails",
                            "succeeds",
                            "dislodged",
                            "bounces",
                        ]
                    ):
                        return line

        return "No description found"

    def _extract_orders(self, content: str) -> Dict[str, List[str]]:
        """Extract orders for each nation."""
        orders = {}
        lines = content.split("\n")
        current_nation = None

        for line in lines:
            line = line.strip()

            # Check for nation header (e.g., "England:")
            nation_match = re.match(
                r"^(England|France|Germany|Austria|Italy|Russia|Turkey):\s*$",
                line,
            )
            if nation_match:
                current_nation = nation_match.group(1)
                orders[current_nation] = []
                continue

            # If we have a current nation and this looks like an order
            if current_nation and line:
                # Skip outcome descriptions
                if not any(
                    keyword in line.lower()
                    for keyword in [
                        "order should",
                        "move fails",
                        "succeeds",
                        "dislodged",
                        "bounces",
                        "will move",
                    ]
                ):
                    # Check if this looks like an order
                    if any(
                        pattern in line
                        for pattern in [
                            "-",
                            "Hold",
                            "Supports",
                            "Convoys",
                            "Build",
                            "Remove",
                        ]
                    ):
                        orders[current_nation].append(line)

        return orders

    def _extract_expected_outcome(self, content: str) -> str:
        """Extract expected outcome description."""
        outcome_keywords = [
            "order should",
            "move fails",
            "succeeds",
            "dislodged",
            "bounces",
            "will move",
            "will not move",
            "fails",
            "advances",
        ]

        lines = content.split("\n")
        outcome_lines = []

        for line in lines:
            line = line.strip()
            if any(keyword in line.lower() for keyword in outcome_keywords):
                outcome_lines.append(line)

        return (
            " ".join(outcome_lines) if outcome_lines else "No outcome specified"
        )

    def _get_category_from_section(self, section_title: str) -> str:
        """Convert section title to category name."""
        category_map = {
            "BASIC CHECKS": "basic_checks",
            "COASTAL ISSUES": "coastal_issues",
            "CIRCULAR MOVEMENT": "circular_movement",
            "SUPPORTS AND DISLODGES": "supports_dislodges",
            "HEAD-TO-HEAD BATTLES": "head_to_head",
            "BELEAGUERED GARRISON": "head_to_head",
            "CONVOYS": "convoys",
            "CONVOYING TO ADJACENT": "adjacent_convoys",
            "RETREATING": "retreating",
            "BUILDING": "building",
            "CIVIL DISORDER": "civil_disorder",
            "DISBANDS": "civil_disorder",
        }

        section_upper = section_title.upper()
        for key, value in category_map.items():
            if key in section_upper:
                return value

        return "unknown"

    def save_to_json(self, test_cases: List[Dict], output_path: Path) -> None:
        """Save test cases to JSON file."""
        print(f"Saving {len(test_cases)} test cases to {output_path}...")

        output_data = {
            "datc_version": "3.0",
            "extracted_by": "Simple DATC Extractor",
            "total_test_cases": len(test_cases),
            "test_cases": test_cases,
        }

        with open(output_path, "w", encoding="utf-8") as f:
            json.dump(output_data, f, indent=2, ensure_ascii=False)

        print(f"Successfully saved test cases to {output_path}")


def main():
    parser = argparse.ArgumentParser(
        description="Extract DATC test cases (simple version)"
    )
    parser.add_argument("html_file", help="DATC HTML file to parse")
    parser.add_argument(
        "--output",
        "-o",
        default="datc_tests.json",
        help="Output JSON file (default: datc_tests.json)",
    )

    args = parser.parse_args()

    html_path = Path(args.html_file)
    if not html_path.exists():
        print(f"Error: HTML file {html_path} not found")
        sys.exit(1)

    extractor = SimpleDATCExtractor()

    # Extract test cases
    test_cases = extractor.extract_from_html(html_path)

    # Save to JSON
    output_path = Path(args.output)
    extractor.save_to_json(test_cases, output_path)

    # Print summary
    print("\n=== Extraction Summary ===")
    print(f"Total test cases: {len(test_cases)}")

    # Group by category
    categories = {}
    for test_case in test_cases:
        category = test_case["category"]
        categories[category] = categories.get(category, 0) + 1

    print("\nBy category:")
    for category, count in sorted(categories.items()):
        print(f"  {category}: {count}")

    # Show first few test cases as examples
    print("\nFirst 3 test cases:")
    for i, test_case in enumerate(test_cases[:3]):
        print(f"  {test_case['id']}: {test_case['name']}")


if __name__ == "__main__":
    main()
