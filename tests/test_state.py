import unittest
from logic.state import apply_unit_movements

class TestState(unittest.TestCase):
    def test_apply_movements(self):
        test_state = {
            "units": {
                "PAR": { "owner_id": "FRA", "type": "army" },
                "MAR": { "owner_id": "FRA", "type": "army" },
                "BRE": { "owner_id": "FRA", "type": "fleet" }
            }
        }
        test_movements = [
            {"from": "BRE", "to": "MAO"},
            {"from": "PAR", "to": "BRE"},
            {"from": "MAR", "to": "SPA"}
        ]

        expected = {
            "units": {
                "BRE": { "owner_id": "FRA", "type": "army" },
                "SPA": { "owner_id": "FRA", "type": "army" },
                "MAO": { "owner_id": "FRA", "type": "fleet" }
            }
        }

        result = apply_unit_movements(test_state, test_movements)
        self.assertEqual(result, expected)

if __name__ == '__main__':
    unittest.main() 


