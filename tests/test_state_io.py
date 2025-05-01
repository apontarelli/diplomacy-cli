import tempfile
import unittest
from pathlib import Path

from diplomacy_cli.core.logic.state import (
    INITIAL_TURN_CODE,
    build_counters,
    build_territory_to_unit,
    load_state,
    start_game,
)
from diplomacy_cli.core.logic.storage import load


class TestStateIO(unittest.TestCase):
    def setUp(self):
        self._tmpdir_obj = tempfile.TemporaryDirectory()
        self.tmpdir = Path(self._tmpdir_obj.name)
        self.game_id = "test_game"

    def tearDown(self):
        self._tmpdir_obj.cleanup()

    def _all_json_files(self, game_id):
        base = self.tmpdir / game_id
        return [
            base / "players.json",
            base / "units.json",
            base / "territory_state.json",
            base / "game.json",
            base / "orders.json",
        ]

    def test_round_trip_identity(self):
        game_id = "io_test"
        start_game(game_id=game_id, save_dir=self.tmpdir)

        for path in self._all_json_files(game_id):
            self.assertTrue(path.exists(), msg=f"{path.name} missing")
            self.assertTrue(path.stat().st_size > 0, msg=f"{path.name} empty")

        reloaded_state, t2u, counters = load_state(game_id, save_dir=self.tmpdir)

        saved_players = reloaded_state["players"]
        saved_units = reloaded_state["units"]
        saved_terr = reloaded_state["territory_state"]
        saved_game = reloaded_state["game"]

        self.assertEqual(saved_players, load(f"{self.tmpdir}/{game_id}/players.json"))
        self.assertEqual(saved_units, load(f"{self.tmpdir}/{game_id}/units.json"))
        self.assertEqual(saved_terr, load(f"{self.tmpdir}/{game_id}/territory_state.json"))
        self.assertEqual(saved_game["game_id"], game_id)
        self.assertEqual(saved_game["turn_code"], INITIAL_TURN_CODE)

        self.assertEqual(t2u, build_territory_to_unit(saved_units))
        self.assertEqual(counters, build_counters(saved_units))

    def test_start_game_overwrite_protection(self):
        game_id = "clobber_me"
        start_game(game_id=game_id, save_dir=self.tmpdir)
        with self.assertRaises(FileExistsError):
            start_game(game_id=game_id)
