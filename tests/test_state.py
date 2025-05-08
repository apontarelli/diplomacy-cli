import unittest

from diplomacy_cli.core.logic.state import (
    apply_unit_movements,
    build_counters,
    build_territory_to_unit,
    build_unit,
    disband_unit,
    eliminate_player,
    set_territory_owner,
)


class TestState(unittest.TestCase):
    maxDiff = None

    def test_build_territory_to_unit(self):
        test_units = {
            "FRA_army_1": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "PAR",
            },
            "FRA_army_2": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "MAR",
            },
            "FRA_fleet_1": {
                "owner_id": "FRA",
                "unit_type": "fleet",
                "territory_id": "BRE",
            },
        }

        expected = {
            "PAR": "FRA_army_1",
            "MAR": "FRA_army_2",
            "BRE": "FRA_fleet_1",
        }

        result = build_territory_to_unit(test_units)
        self.assertEqual(result, expected)

    def test_apply_movements(self):
        test_units = {
            "FRA_army_1": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "PAR",
            },
            "FRA_army_2": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "MAR",
            },
            "FRA_fleet_1": {
                "owner_id": "FRA",
                "unit_type": "fleet",
                "territory_id": "BRE",
            },
        }
        test_movements = [
            {"from": "BRE", "to": "MAO"},
            {"from": "PAR", "to": "BRE"},
            {"from": "MAR", "to": "SPA"},
        ]

        expected = {
            "FRA_army_1": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "BRE",
            },
            "FRA_army_2": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "SPA",
            },
            "FRA_fleet_1": {
                "owner_id": "FRA",
                "unit_type": "fleet",
                "territory_id": "MAO",
            },
        }

        expected_t2u = {
            "BRE": "FRA_army_1",
            "SPA": "FRA_army_2",
            "MAO": "FRA_fleet_1",
        }

        territory_to_unit = build_territory_to_unit(test_units)
        result, result_t2u = apply_unit_movements(
            test_units, territory_to_unit, test_movements
        )
        self.assertEqual(result, expected)
        self.assertEqual(result_t2u, expected_t2u)

    def test_disband_unit(self):
        test_units = {
            "FRA_army_1": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "PAR",
            },
            "FRA_army_2": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "MAR",
            },
            "FRA_fleet_1": {
                "owner_id": "FRA",
                "unit_type": "fleet",
                "territory_id": "BRE",
            },
        }

        expected = {
            "FRA_army_2": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "MAR",
            },
            "FRA_fleet_1": {
                "owner_id": "FRA",
                "unit_type": "fleet",
                "territory_id": "BRE",
            },
        }

        expected_t2u = {"MAR": "FRA_army_2", "BRE": "FRA_fleet_1"}

        territory_to_unit = build_territory_to_unit(test_units)
        result, result_t2u = disband_unit(test_units, territory_to_unit, "PAR")
        self.assertEqual(result, expected)
        self.assertEqual(result_t2u, expected_t2u)

    def test_build_unit(self):
        test_units = {
            "FRA_army_1": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "PAR",
            },
            "FRA_army_2": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "PIC",
            },
            "FRA_fleet_1": {
                "owner_id": "FRA",
                "unit_type": "fleet",
                "territory_id": "MAO",
            },
        }

        expected = {
            "FRA_army_1": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "PAR",
            },
            "FRA_army_2": {
                "owner_id": "FRA",
                "unit_type": "army",
                "territory_id": "PIC",
            },
            "FRA_fleet_1": {
                "owner_id": "FRA",
                "unit_type": "fleet",
                "territory_id": "MAO",
            },
            "FRA_fleet_2": {
                "owner_id": "FRA",
                "unit_type": "fleet",
                "territory_id": "BRE",
            },
            "GER_army_1": {
                "owner_id": "GER",
                "unit_type": "army",
                "territory_id": "MUN",
            },
        }

        expected_t2u = {
            "PAR": "FRA_army_1",
            "PIC": "FRA_army_2",
            "MAO": "FRA_fleet_1",
            "BRE": "FRA_fleet_2",
            "MUN": "GER_army_1",
        }

        expected_counters = {"FRA_army": 2, "FRA_fleet": 2, "GER_army": 1}

        territory_to_unit = build_territory_to_unit(test_units)
        counters = build_counters(test_units)
        result, result_t2u, result_counters = build_unit(
            test_units, territory_to_unit, counters, "BRE", "fleet", "FRA"
        )
        result, result_t2u, result_counters = build_unit(
            result, result_t2u, result_counters, "MUN", "army", "GER"
        )
        self.assertEqual(result, expected)
        self.assertEqual(result_t2u, expected_t2u)
        self.assertEqual(result_counters, expected_counters)

    def test_set_territory_owner(self):
        test_territory_state = {
            "LON": {"owner_id": "ENG"},
            "LVP": {"owner_id": "ENG"},
            "EDI": {"owner_id": "ENG"},
        }

        expected = {
            "LON": {"owner_id": "FRA"},
            "LVP": {"owner_id": "ENG"},
            "EDI": {"owner_id": "RUS"},
        }

        result = set_territory_owner(test_territory_state, "LON", "FRA")
        result = set_territory_owner(result, "EDI", "RUS")
        self.assertEqual(result, expected)

    def test_eliminate_player(self):
        test_players = {
            "ENG": {"status": "active"},
            "AUS": {"status": "active"},
            "GER": {"status": "active"},
            "FRA": {"status": "active"},
            "ITA": {"status": "active"},
            "TUR": {"status": "active"},
            "RUS": {"status": "active"},
        }

        expected = {
            "ENG": {"status": "eliminated"},
            "AUS": {"status": "active"},
            "GER": {"status": "active"},
            "FRA": {"status": "active"},
            "ITA": {"status": "active"},
            "TUR": {"status": "active"},
            "RUS": {"status": "active"},
        }

        result = eliminate_player(test_players, "ENG")
        self.assertEqual(result, expected)


if __name__ == "__main__":
    unittest.main()
