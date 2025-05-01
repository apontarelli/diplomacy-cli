import json
import pathlib
import unittest

from jsonschema import Draft202012Validator


class BoardSchemaTest(unittest.TestCase):
    def test_territories_valid(self):
        base = pathlib.Path("src/diplomacy_cli/data/classic/world")
        schema = json.loads((base / "board.schema.json").read_text())
        data = json.loads((base / "territories.json").read_text())
        Draft202012Validator(schema).validate(data)  # raises on failure
