#!/usr/bin/env python3
"""
Extract DATC test cases from the webdiplomacy.net HTML document.
Creates Go test files organized by category.
"""

import re
import urllib.request
from pathlib import Path
from typing import List, Dict, Tuple
from dataclasses import dataclass

@dataclass
class TestCase:
    id: str
    title: str
    description: str
    orders: List[str]
    expected_result: str
    category: str

def fetch_datc_html() -> str:
    """Fetch the DATC HTML from webdiplomacy.net"""
    url = "https://webdiplomacy.net/doc/DATC_v3_0.html"
    
    # Try to read from local file first
    local_file = Path("scripts/DATC_v3_0.html")
    if local_file.exists():
        print(f"Using local file: {local_file}")
        return local_file.read_text(encoding='utf-8')
    
    # Otherwise fetch from web with user agent
    req = urllib.request.Request(url)
    req.add_header('User-Agent', 'Mozilla/5.0 (Diplomacy CLI Test Extractor)')
    
    try:
        with urllib.request.urlopen(req) as response:
            html = response.read().decode('utf-8')
            # Save for future use
            local_file.write_text(html, encoding='utf-8')
            print(f"Downloaded and saved to: {local_file}")
            return html
    except Exception as e:
        print(f"Failed to fetch from web: {e}")
        print(f"Please manually download {url} and save as {local_file}")
        raise

def extract_test_cases(html: str) -> List[TestCase]:
    """Extract all test cases from the HTML"""
    test_cases = []
    
    # Find all h4 tags that contain test cases
    h4_pattern = r'<h4><a name="([^"]+)">([^<]+)</a></h4>'
    h4_matches = re.finditer(h4_pattern, html)
    
    for match in h4_matches:
        test_id = match.group(1)
        title = match.group(2).strip()
        
        # Skip non-test case h4s (like section headers)
        if not re.match(r'6\.[A-J]\.\d+', test_id):
            continue
            
        # Extract category from test_id (e.g., "6.A.1" -> "A")
        category = test_id.split('.')[1]
        
        # Find the content between this h4 and the next h4 or h3
        start_pos = match.end()
        next_header = re.search(r'<h[34]', html[start_pos:])
        if next_header:
            end_pos = start_pos + next_header.start()
        else:
            end_pos = len(html)
            
        content = html[start_pos:end_pos]
        
        # Extract description (first paragraph)
        desc_match = re.search(r'<p>([^<]+(?:<[^>]+>[^<]*</[^>]+>[^<]*)*)</p>', content)
        description = desc_match.group(1) if desc_match else ""
        description = re.sub(r'<[^>]+>', '', description).strip()
        
        # Extract orders (content within <pre> tags)
        orders = []
        pre_matches = re.finditer(r'<pre>(.*?)</pre>', content, re.DOTALL)
        for pre_match in pre_matches:
            pre_content = pre_match.group(1).strip()
            # Split by country and clean up
            country_orders = re.split(r'\n(?=[A-Z][a-z]+:)', pre_content)
            for country_block in country_orders:
                if ':' in country_block:
                    orders.append(country_block.strip())
        
        # Extract expected result (look for key phrases)
        expected_result = ""
        if "should fail" in content.lower() or "fails" in content.lower():
            expected_result = "fails"
        elif "succeeds" in content.lower() or "will move" in content.lower():
            expected_result = "succeeds"
        elif "bounce" in content.lower():
            expected_result = "bounces"
        elif "dislodged" in content.lower():
            expected_result = "dislodged"
        
        test_cases.append(TestCase(
            id=test_id,
            title=title,
            description=description,
            orders=orders,
            expected_result=expected_result,
            category=category
        ))
    
    return test_cases

def generate_go_test_file(category: str, test_cases: List[TestCase]) -> str:
    """Generate a Go test file for a specific category"""
    
    category_names = {
        'A': 'BasicChecks',
        'B': 'CoastalIssues', 
        'C': 'CircularMovement',
        'D': 'SupportsAndDislodges',
        'E': 'HeadToHeadBattles',
        'F': 'Convoys',
        'G': 'ConvoyingToAdjacent',
        'H': 'Retreating',
        'I': 'Building',
        'J': 'CivilDisorder'
    }
    
    category_name = category_names.get(category, f'Category{category}')
    
    content = f'''package api

import (
	"testing"
)

// DATC Test Cases - Category {category}: {category_name}
// Generated from https://webdiplomacy.net/doc/DATC_v3_0.html

'''
    
    for test_case in test_cases:
        # Convert test ID to valid Go function name
        func_name = f"TestDATC_{test_case.id.replace('.', '_')}"
        
        content += f'''func {func_name}(t *testing.T) {{
	// {test_case.title}
	// {test_case.description}
	
	t.Skip("DATC test case not yet implemented")
	
	// TODO: Implement test case
	// Orders:
'''
        
        for order in test_case.orders:
            # Split multi-line orders and add each line as a separate comment
            lines = order.replace('\r', '').split('\n')
            for line in lines:
                line = line.strip()
                if line:
                    content += f'\t// {line}\\n'
            
        if test_case.expected_result:
            content += f'	// Expected: {test_case.expected_result}\n'
            
        content += '}\n\n'
    
    return content

def main():
    """Main function to extract and generate test files"""
    print("Fetching DATC HTML...")
    html = fetch_datc_html()
    
    print("Extracting test cases...")
    test_cases = extract_test_cases(html)
    
    print(f"Found {len(test_cases)} test cases")
    
    # Group by category
    by_category: Dict[str, List[TestCase]] = {}
    for test_case in test_cases:
        if test_case.category not in by_category:
            by_category[test_case.category] = []
        by_category[test_case.category].append(test_case)
    
    # Create output directory
    output_dir = Path("backend/internal/api")
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Generate test files for each category
    for category, cases in by_category.items():
        print(f"Generating tests for category {category}: {len(cases)} cases")
        
        go_content = generate_go_test_file(category, cases)
        
        filename = f"datc_{category.lower()}_test.go"
        filepath = output_dir / filename
        
        with open(filepath, 'w') as f:
            f.write(go_content)
        
        print(f"  -> {filepath}")
    
    print(f"\nGenerated {len(by_category)} test files with {len(test_cases)} total test cases")
    print("\nNext steps:")
    print("1. Review the generated test files")
    print("2. Implement the test cases by adding actual game setup and assertions")
    print("3. Run tests with: go test ./backend/internal/api/...")

if __name__ == "__main__":
    main()