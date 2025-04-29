import os
import json
import shutil

DEFAULT_SAVES_DIR = "data/saves"


def load(path):
    with open(path) as f:
        return json.load(f)


def save(data, path):
    with open(path, "w") as f:
        json.dump(data, f, indent=2)


def delete_game(game_id):
    game_path = f"data/saves/{game_id}"
    shutil.rmtree(game_path)


def list_saved_games(saves_dir=DEFAULT_SAVES_DIR):
    return os.listdir(saves_dir)
