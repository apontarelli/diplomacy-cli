from pathlib import Path

from diplomacy_cli.core.logic.storage import load, save


def test_storage_interface(tmp_path: Path):
    games_root = tmp_path / "data" / "games"
    games_root.mkdir(parents=True)

    test_data = {"test": "data"}
    game_dir = games_root / "test_game"
    game_file = game_dir / "game.json"

    save(test_data, game_file)
    assert game_file.exists(), "Expected save() to create the game.json file"

    loaded = load(game_file)
    assert loaded == test_data
