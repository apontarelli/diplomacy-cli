import json
import os
import shutil
from importlib import resources
from pathlib import Path
from typing import Any

from platformdirs import user_data_path

type Pathish = str | os.PathLike[str]

DEFAULT_GAMES_DIR = (
    Path(os.getenv("DCLI_GAMES_DIR", user_data_path("diplomacy-cli"))) / "games"
)
DEFAULT_GAMES_DIR.mkdir(parents=True, exist_ok=True)


def load(source: str | bytes | Pathish) -> dict[str, Any]:
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
        raw = json.load(f)
    if not isinstance(raw, dict):
        raise ValueError(f"{source!r} is not a JSON object")
    return raw


def load_variant_json(
    variant: str, submodule: str, filename: str
) -> dict | list:
    pkg = f"diplomacy_cli.data.{variant}.{submodule}"
    path = resources.files(pkg).joinpath(filename)
    return load(path.read_text(encoding="utf-8"))


def save(data: dict | list, path: Pathish) -> None:
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)


def delete_game(game_id: str) -> None:
    shutil.rmtree(DEFAULT_GAMES_DIR / game_id, ignore_errors=False)


def list_games(games_dir: Pathish = DEFAULT_GAMES_DIR) -> list[str]:
    games_dir = Path(games_dir)
    return [p.name for p in games_dir.iterdir() if p.is_dir()]
