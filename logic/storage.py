import os
import json

DEFAULT_SAVES_DIR = "data/saves"

def load(path):
    print("Loading game")
    with open(path) as f:
        return json.load(f)

def save(data, path):
    print("Saving game")
    with open(path, "w") as f:
        json.dump(data, f, indent=2)

def list_saved_games(saves_dir=DEFAULT_SAVES_DIR):
    return os.listdir(saves_dir)
