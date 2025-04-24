import unittest
from logic.state import apply_unit_movements, disband_unit, build_unit, set_territory_owner, eliminate_player

class TestState(unittest.TestCase):
    def test_apply_movements(self):
        test_units = {
            "PAR": { "owner_id": "FRA", "type": "army" },
            "MAR": { "owner_id": "FRA", "type": "army" },
            "BRE": { "owner_id": "FRA", "type": "fleet" }
        }
        test_movements = [
            {"from": "BRE", "to": "MAO"},
            {"from": "PAR", "to": "BRE"},
            {"from": "MAR", "to": "SPA"}
        ]

        expected = {
            "BRE": { "owner_id": "FRA", "type": "army" },
            "SPA": { "owner_id": "FRA", "type": "army" },
            "MAO": { "owner_id": "FRA", "type": "fleet" }
        }

        result = apply_unit_movements(test_units, test_movements)
        self.assertEqual(result, expected)

    def test_disband_unit(self):
        test_units = {
            "PAR": { "owner_id": "FRA", "type": "army" },
            "MAR": { "owner_id": "FRA", "type": "army" },
            "BRE": { "owner_id": "FRA", "type": "fleet" }
        }

        expected = {
            "MAR": { "owner_id": "FRA", "type": "army" },
            "BRE": { "owner_id": "FRA", "type": "fleet" }
        }
        
        result = disband_unit(test_units, "PAR")
        self.assertEqual(result, expected)

    def test_build_unit(self):
        test_units = {
            "PAR": { "owner_id": "FRA", "type": "army" },
            "PIC": { "owner_id": "FRA", "type": "army" },
            "MAO": { "owner_id": "FRA", "type": "fleet" }
        }

        expected = {
            "PAR": { "owner_id": "FRA", "type": "army" },
            "PIC": { "owner_id": "FRA", "type": "army" },
            "MAO": { "owner_id": "FRA", "type": "fleet" },
            "BRE": { "owner_id": "FRA", "type": "fleet" },
            "MUN": { "owner_id": "GER", "type": "army" }
        }

        result = build_unit(test_units, "BRE", "fleet", "FRA")
        result = build_unit(result, "MUN", "army", "GER")
        self.assertEqual(result, expected)

    def test_set_territory_owner(self):
        test_territory_state = {
            "LON": { "owner_id": "ENG" },
            "LVP": { "owner_id": "ENG" },
            "EDI": { "owner_id": "ENG" }
        }

        expected = {
            "LON": { "owner_id": "FRA" },
            "LVP": { "owner_id": "ENG" },
            "EDI": { "owner_id": "RUS" }
        }

        result = set_territory_owner(test_territory_state, "LON", "FRA")
        result = set_territory_owner(result, "EDI", "RUS")
        self.assertEqual(result, expected)

    def test_eliminate_player(self):
        test_players = {
            "ENG": { "status": "active" },
            "AUS": { "status": "active" },
            "GER": { "status": "active" },
            "FRA": { "status": "active" },
            "ITA": { "status": "active" },
            "TUR": { "status": "active" },
            "RUS": { "status": "active" }
        }

        expected = {
            "ENG": { "status": "eliminated" },
            "AUS": { "status": "active" },
            "GER": { "status": "active" },
            "FRA": { "status": "active" },
            "ITA": { "status": "active" },
            "TUR": { "status": "active" },
            "RUS": { "status": "active" }
        }



        result = eliminate_player(test_players, "ENG")
        self.assertEqual(result, expected)

if __name__ == '__main__':
    unittest.main() 


