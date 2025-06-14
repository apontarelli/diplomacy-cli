import os
import shutil
from pathlib import Path
from typing import NamedTuple
from diplomacy_cli.core.logic.turn_code import format_turn_code, Season, Phase

from platformdirs import user_data_path

DEFAULT_GAMES_DIR = (
    Path(os.getenv("DCLI_GAMES_DIR", user_data_path("diplomacy-cli"))) / "games"
)
DEFAULT_GAMES_DIR.mkdir(parents=True, exist_ok=True)


class GamePaths(NamedTuple):
    root: Path
    game_id: str


def game_dir(paths: GamePaths) -> Path:
    return paths.root / paths.game_id


def reports_dir(paths: GamePaths) -> Path:
    return game_dir(paths) / "reports"


def ensure_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def report_path(
    paths: GamePaths,
    year: int,
    season: Season,
    phase: Phase,
    create_dir: bool = True,
) -> Path:
    rpt_dir = reports_dir(paths)
    if create_dir:
        ensure_dir(rpt_dir)
    code = format_turn_code(year, season, phase)
    return rpt_dir / f"{code}_report.json"


def game_meta_path(paths: GamePaths) -> Path:
    return game_dir(paths) / "game.json"


def players_path(paths: GamePaths) -> Path:
    return game_dir(paths) / "players.json"


def units_path(paths: GamePaths) -> Path:
    return game_dir(paths) / "units.json"


def orders_path(paths: GamePaths) -> Path:
    return game_dir(paths) / "orders.json"


def territory_state_path(paths: GamePaths) -> Path:
    return game_dir(paths) / "territory_state.json"


def delete_game(game_id: str) -> None:
    dir_path = DEFAULT_GAMES_DIR / game_id
    if not dir_path.is_dir():
        raise FileNotFoundError(
            f"No game directory for '{game_id}' at {DEFAULT_GAMES_DIR!r}"
        )
    shutil.rmtree(dir_path)


def list_game_ids(games_dir: Path = DEFAULT_GAMES_DIR) -> list[str]:
    games_dir = Path(games_dir)
    return [p.name for p in games_dir.iterdir() if p.is_dir()]
