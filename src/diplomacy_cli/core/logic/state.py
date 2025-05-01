from __future__ import annotations

import os
from importlib import resources
from pathlib import Path
from typing import Any

from .storage import DEFAULT_SAVES_DIR, load, save
from .turn_code import INITIAL_TURN_CODE

# --------------------------------------------------------------------------- #
# Types
# --------------------------------------------------------------------------- #
GameDict = dict[str, Any]  # coarse for now; refine later
TerritoryToUnit = dict[str, str]
Counters = dict[str, int]
StateTuple = tuple[GameDict, TerritoryToUnit, Counters]


# --------------------------------------------------------------------------- #
# Internal helpers
# --------------------------------------------------------------------------- #
def _variant_resource(variant: str, filename: str) -> str:
    pkg = f"diplomacy_cli.data.{variant}.start"
    with resources.files(pkg).joinpath(filename).open("r", encoding="utf-8") as fp:
        return fp.read()


def start_game(
    *,
    variant: str = "classic",
    game_id: str = "new_game",
    save_dir: str | os.PathLike[str] | None = None,
) -> None:
    save_root = Path(save_dir or DEFAULT_SAVES_DIR)
    save_path = save_root / game_id
    if save_path.exists():
        raise FileExistsError(f"Save directory '{save_path}' already exists.")
    save_path.mkdir(parents=True, exist_ok=True)

    starting_units = load(_variant_resource(variant, "starting_units.json"))
    starting_ownerships = load(_variant_resource(variant, "starting_ownerships.json"))
    starting_players = load(_variant_resource(variant, "starting_players.json"))

    state: GameDict = {
        "players": {},
        "units": {},
        "territory_state": {},
        "orders": {},
        "game": {
            "game_id": game_id,
            "variant": variant,
            "turn_code": INITIAL_TURN_CODE,
            "status": "active",
        },
    }

    for player in starting_players:
        state["players"][player["nation_id"]] = {"status": player["status"]}

    for o in starting_ownerships:
        territory_id = o["territory_id"]
        owner_id = o["owner_id"]
        state["territory_state"] = set_territory_owner(
            state["territory_state"], territory_id, owner_id
        )

    counters: Counters = {}
    territory_to_unit: TerritoryToUnit = {}
    for u in starting_units:
        state["units"], territory_to_unit, counters = build_unit(
            state["units"],
            territory_to_unit,
            counters,
            u["location_id"],
            u["unit_type"],
            u["owner_id"],
        )

    save(state["players"], save_path / "players.json")
    save(state["units"], save_path / "units.json")
    save(state["territory_state"], save_path / "territory_state.json")
    save(state["game"], save_path / "game.json")
    save(state["orders"], save_path / "orders.json")

    print(f"Game {game_id} created successfully!")


def load_state(game_id: str, *, save_dir: str | os.PathLike[str] | None = None) -> StateTuple:
    """
    Return

        (full_state_dict, territory_to_unit_index, counters_index)
    """
    save_root = Path(save_dir or DEFAULT_SAVES_DIR)
    save_path = save_root / game_id

    state: GameDict = {
        "game": load(save_path / "game.json"),
        "players": load(save_path / "players.json"),
        "territory_state": load(save_path / "territory_state.json"),
        "units": load(save_path / "units.json"),
        "orders": load(save_path / "orders.json"),
    }

    territory_to_unit = build_territory_to_unit(state["units"])
    counters = build_counters(state["units"])

    return state, territory_to_unit, counters


def build_territory_to_unit(units: dict[str, dict[str, Any]]) -> TerritoryToUnit:
    return {unit["territory_id"]: unit_id for unit_id, unit in units.items()}


def build_counters(units: dict[str, Any]) -> Counters:
    counters: Counters = {}
    for unit_id in units:
        parts = unit_id.split("_")
        if len(parts) != 3:
            continue
        owner, unit_type, num = parts
        key = f"{owner}_{unit_type}"
        counters[key] = max(counters.get(key, 0), int(num))
    return counters


def apply_unit_movements(
    units: dict[str, Any],
    territory_to_unit: TerritoryToUnit,
    movements: list[dict[str, str]],
) -> tuple[dict[str, Any], TerritoryToUnit]:
    for move in movements:
        from_ = move["from"]
        to = move["to"]
        unit_id = territory_to_unit[from_]
        units[unit_id]["territory_id"] = to
        territory_to_unit[to] = territory_to_unit.pop(from_)
    return units, territory_to_unit


def disband_unit(
    units: dict[str, Any],
    territory_to_unit: TerritoryToUnit,
    territory_id: str,
) -> tuple[dict[str, Any], TerritoryToUnit]:
    unit_id = territory_to_unit.pop(territory_id)
    units.pop(unit_id)
    return units, territory_to_unit


def build_unit(
    units: dict[str, Any],
    territory_to_unit: TerritoryToUnit,
    counters: Counters,
    territory_id: str,
    unit_type: str,
    owner_id: str,
) -> tuple[dict[str, Any], TerritoryToUnit, Counters]:
    key = f"{owner_id}_{unit_type}"
    next_num = counters.get(key, 0) + 1
    unit_id = f"{key}_{next_num}"

    units[unit_id] = {
        "unit_type": unit_type,
        "owner_id": owner_id,
        "territory_id": territory_id,
    }
    territory_to_unit[territory_id] = unit_id
    counters[key] = next_num

    return units, territory_to_unit, counters


def set_territory_owner(
    territory_state: dict[str, dict[str, str]], territory_id: str, owner_id: str
) -> dict[str, dict[str, str]]:
    territory_state[territory_id] = {"owner_id": owner_id}
    return territory_state


def eliminate_player(players: dict[str, dict[str, str]], player_id: str) -> dict[str, Any]:
    players[player_id] = {"status": "eliminated"}
    return players
