import os
import shutil
import tempfile
import unittest

from diplomacy_cli.core.logic.storage import list_games, load, save


class TestStorage(unittest.TestCase):
    def setUp(self):
        self.test_dir = tempfile.mkdtemp()
        self.games_dir = os.path.join(self.test_dir, "data", "games")
        os.makedirs(self.games_dir)

    def tearDown(self):
        shutil.rmtree(self.test_dir)

    def test_storage_interface(self):
        test_data = {"test": "data"}
        game_dir = os.path.join(self.games_dir, "test_game")
        game_path = os.path.join(game_dir, "game.json")

        save(test_data, game_path)
        self.assertTrue(os.path.exists(game_path))

        loaded_data = load(game_path)
        self.assertEqual(loaded_data, test_data)

        games = list_games(self.games_dir)
        self.assertIn("test_game", games)


if __name__ == "__main__":
    unittest.main()
