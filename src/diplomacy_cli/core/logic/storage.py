from __future__ import annotations

import json
import os
import shutil
from pathlib import Path

from platformdirs import user_data_path

type Pathish = str | os.PathLike[str]

DEFAULT_SAVES_DIR = Path(os.getenv("DCLI_SAVES_DIR", user_data_path("diplomacy-cli"))) / "saves"
DEFAULT_SAVES_DIR.mkdir(parents=True, exist_ok=True)


def load(source: str | bytes | Pathish) -> dict | list:
    if isinstance(source, str):
        s = source.lstrip()
        if s.startswith(("{", "[")):
            return json.loads(s)
        source = Path(s)

    elif isinstance(source, (bytes | bytearray)):
        s = source.lstrip()
        if s.startswith((b"{", b"[")):
            return json.loads(s)
        source = Path(s.decode())

    else:
        source = Path(source)

    with source.open("r", encoding="utf-8") as f:
        return json.load(f)


def save(data: dict | list, path: Pathish) -> None:
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)


def delete_game(game_id: str) -> None:
    shutil.rmtree(DEFAULT_SAVES_DIR / game_id, ignore_errors=False)


def list_saved_games(saves_dir: Pathish = DEFAULT_SAVES_DIR) -> list[str]:
    saves_dir = Path(saves_dir)
    return [p.name for p in saves_dir.iterdir() if p.is_dir()]
