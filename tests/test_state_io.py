
import os
import shutil
import tempfile
import unittest
from pathlib import Path

from ..logic import state

class TestStateIO(unittest.TestCase):

    def setUp(self):
        self._tmpdir = tempfile.mkdtemp()
        self._old_dir = state.DEFAULT_SAVES_DIR
        state.DEFAULT_SAVES_DIR = self._tmpdir

    def tearDown(self):
        state.DEFAULT_SAVES_DIR = self._old_dir
        shutil.rmtree(self._tmpdir)

    def _all_json_files(self, game_id):
        base = Path(self._tmpdir) / game_id
        return [
            base / "players.json",
            base / "units.json",
            base / "territory_state.json",
            base / "game.json",
            base / "orders.json",
        ]
    def test_round_trip_identity(self):
        game_id = "io_test"
        state.start_game(game_id=game_id)          # uses classic variant by default

        for path in self._all_json_files(game_id):
            self.assertTrue(path.exists(), msg=f"{path.name} missing")
            self.assertTrue(path.stat().st_size > 0, msg=f"{path.name} empty")

        reloaded_state, t2u, counters = state.load_state(game_id)

        saved_players = reloaded_state["players"]
        saved_units   = reloaded_state["units"]
        saved_terr    = reloaded_state["territory_state"]
        saved_game    = reloaded_state["game"]

        self.assertEqual(saved_players,
                         state.load(f"{self._tmpdir}/{game_id}/players.json"))
        self.assertEqual(saved_units,
                         state.load(f"{self._tmpdir}/{game_id}/units.json"))
        self.assertEqual(saved_terr,
                         state.load(f"{self._tmpdir}/{game_id}/territory_state.json"))
        self.assertEqual(saved_game["game_id"], game_id)
        self.assertEqual(saved_game["turn_code"], state.INITIAL_TURN_CODE)

        self.assertEqual(t2u, state.build_territory_to_unit(saved_units))
        self.assertEqual(counters, state.build_counters(saved_units))

    def test_start_game_overwrite_protection(self):
        game_id = "clobber_me"
        state.start_game(game_id=game_id)
        with self.assertRaises(FileExistsError):
            state.start_game(game_id=game_id)
