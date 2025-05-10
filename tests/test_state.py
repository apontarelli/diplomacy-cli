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
            "fra_army_1": {
                "id": "fra_army_1",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "par",
            },
            "fra_army_2": {
                "id": "fra_army_2",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "mar",
            },
            "fra_fleet_1": {
                "id": "fra_fleet_1",
                "owner_id": "fra",
                "unit_type": "fleet",
                "territory_id": "bre",
            },
        }

        expected = {
            "par": "fra_army_1",
            "mar": "fra_army_2",
            "bre": "fra_fleet_1",
        }

        result = build_territory_to_unit(test_units)
        self.assertEqual(result, expected)

    def test_apply_movements(self):
        test_units = {
            "fra_army_1": {
                "id": "fra_army_1",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "par",
            },
            "fra_army_2": {
                "id": "fra_army_2",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "mar",
            },
            "fra_fleet_1": {
                "id": "fra_fleet_1",
                "owner_id": "fra",
                "unit_type": "fleet",
                "territory_id": "bre",
            },
        }
        test_movements = [
            {"from": "bre", "to": "mao"},
            {"from": "par", "to": "bre"},
            {"from": "mar", "to": "spa"},
        ]

        expected = {
            "fra_army_1": {
                "id": "fra_army_1",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "bre",
            },
            "fra_army_2": {
                "id": "fra_army_2",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "spa",
            },
            "fra_fleet_1": {
                "id": "fra_fleet_1",
                "owner_id": "fra",
                "unit_type": "fleet",
                "territory_id": "mao",
            },
        }

        expected_t2u = {
            "bre": "fra_army_1",
            "spa": "fra_army_2",
            "mao": "fra_fleet_1",
        }

        territory_to_unit = build_territory_to_unit(test_units)
        result, result_t2u = apply_unit_movements(
            test_units, territory_to_unit, test_movements
        )
        self.assertEqual(result, expected)
        self.assertEqual(result_t2u, expected_t2u)

    def test_disband_unit(self):
        test_units = {
            "fra_army_1": {
                "id": "fra_army_1",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "par",
            },
            "fra_army_2": {
                "id": "fra_army_2",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "mar",
            },
            "fra_fleet_1": {
                "id": "fra_fleet_1",
                "owner_id": "fra",
                "unit_type": "fleet",
                "territory_id": "bre",
            },
        }

        expected = {
            "fra_army_2": {
                "id": "fra_army_2",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "mar",
            },
            "fra_fleet_1": {
                "id": "fra_fleet_1",
                "owner_id": "fra",
                "unit_type": "fleet",
                "territory_id": "bre",
            },
        }

        expected_t2u = {"mar": "fra_army_2", "bre": "fra_fleet_1"}

        territory_to_unit = build_territory_to_unit(test_units)
        result, result_t2u = disband_unit(test_units, territory_to_unit, "par")
        self.assertEqual(result, expected)
        self.assertEqual(result_t2u, expected_t2u)

    def test_build_unit(self):
        test_units = {
            "fra_army_1": {
                "id": "fra_army_1",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "gas",
            },
            "fra_army_2": {
                "id": "fra_army_2",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "spa",
            },
            "fra_fleet_1": {
                "id": "fra_fleet_1",
                "owner_id": "fra",
                "unit_type": "fleet",
                "territory_id": "mao",
            },
        }

        expected = {
            "fra_army_1": {
                "id": "fra_army_1",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "gas",
            },
            "fra_army_2": {
                "id": "fra_army_2",
                "owner_id": "fra",
                "unit_type": "army",
                "territory_id": "spa",
            },
            "fra_fleet_1": {
                "id": "fra_fleet_1",
                "owner_id": "fra",
                "unit_type": "fleet",
                "territory_id": "mao",
            },
            "fra_fleet_2": {
                "id": "fra_fleet_2",
                "owner_id": "fra",
                "unit_type": "fleet",
                "territory_id": "bre",
            },
            "ger_army_1": {
                "id": "ger_army_1",
                "owner_id": "ger",
                "unit_type": "army",
                "territory_id": "mun",
            },
        }

        expected_t2u = {
            "gas": "fra_army_1",
            "spa": "fra_army_2",
            "mao": "fra_fleet_1",
            "bre": "fra_fleet_2",
            "mun": "ger_army_1",
        }

        expected_counters = {"fra_army": 2, "fra_fleet": 2, "ger_army": 1}

        territory_to_unit = build_territory_to_unit(test_units)
        counters = build_counters(test_units)
        result, result_t2u, result_counters = build_unit(
            test_units, territory_to_unit, counters, "bre", "fleet", "fra"
        )
        result, result_t2u, result_counters = build_unit(
            result, result_t2u, result_counters, "mun", "army", "ger"
        )
        self.assertEqual(result, expected)
        self.assertEqual(result_t2u, expected_t2u)
        self.assertEqual(result_counters, expected_counters)

    def test_set_territory_owner(self):
        test_territory_state = {
            "lon": {"territory_id": "lon", "owner_id": "eng"},
            "lvp": {"territory_id": "lvp", "owner_id": "eng"},
            "edi": {"territory_id": "edi", "owner_id": "eng"},
        }

        expected = {
            "lon": {"territory_id": "lon", "owner_id": "fra"},
            "lvp": {"territory_id": "lvp", "owner_id": "eng"},
            "edi": {"territory_id": "edi", "owner_id": "rus"},
        }

        result = set_territory_owner(test_territory_state, "lon", "fra")
        result = set_territory_owner(result, "edi", "rus")
        self.assertEqual(result, expected)

    def test_eliminate_player(self):
        test_players = {
            "eng": {"nation_id": "eng", "status": "active"},
            "aus": {"nation_id": "aus", "status": "active"},
            "ger": {"nation_id": "ger", "status": "active"},
            "fra": {"nation_id": "fra", "status": "active"},
            "ita": {"nation_id": "ita", "status": "active"},
            "tur": {"nation_id": "tur", "status": "active"},
            "rus": {"nation_id": "rus", "status": "active"},
        }

        expected = {
            "eng": {"nation_id": "eng", "status": "eliminated"},
            "aus": {"nation_id": "aus", "status": "active"},
            "ger": {"nation_id": "ger", "status": "active"},
            "fra": {"nation_id": "fra", "status": "active"},
            "ita": {"nation_id": "ita", "status": "active"},
            "tur": {"nation_id": "tur", "status": "active"},
            "rus": {"nation_id": "rus", "status": "active"},
        }

        result = eliminate_player(test_players, "eng")
        self.assertEqual(result, expected)


if __name__ == "__main__":
    unittest.main()
