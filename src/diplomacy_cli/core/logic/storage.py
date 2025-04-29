import os
from pathlib import Path
from platformdirs import user_data_path
import json
import shutil

DEFAULT_SAVES_DIR = Path(
    os.getenv("DCLI_SAVES_DIR", user_data_path("diplomacy-cli"))
) / "saves"
DEFAULT_SAVES_DIR.mkdir(parents=True, exist_ok=True)

def load(source):
    if isinstance(source, (str, bytes)) and (
        source.lstrip().startswith("{") or source.lstrip().startswith("[")
    ):
        return json.loads(source)
    with open(source) as f:
        return json.load(f)


def save(data, path):
    with open(path, "w") as f:
        json.dump(data, f, indent=2)


def delete_game(game_id):
    game_path = f"data/saves/{game_id}"
    shutil.rmtree(game_path)


def list_saved_games(saves_dir=DEFAULT_SAVES_DIR):
    return os.listdir(saves_dir)
