from __future__ import annotations

import os
from pathlib import Path
from typing import Any

from .schema import Counters, GameState, LoadedState, TerritoryToUnit
from .storage import DEFAULT_GAMES_DIR, load, load_variant_json, save
from .turn_code import INITIAL_TURN_CODE


def start_game(
    *,
    variant: str = "classic",
    game_id: str = "new_game",
    save_dir: str | os.PathLike[str] | None = None,
) -> GameState:
    save_root = Path(save_dir or DEFAULT_GAMES_DIR)
    save_path = save_root / game_id
    if save_path.exists():
        raise FileExistsError(f"Save directory '{save_path}' already exists.")
    save_path.mkdir(parents=True)

    starting_units = load_variant_json(variant, "start", "starting_units.json")
    starting_ownerships = load_variant_json(
        variant, "start", "starting_ownerships.json"
    )
    starting_players = load_variant_json(
        variant, "start", "starting_players.json"
    )

    players = {
        p["nation_id"]: {"nation_id": p["nation_id"], "status": p["status"]}
        for p in starting_players
    }
    game_meta = {
        "game_id": game_id,
        "variant": variant,
        "turn_code": INITIAL_TURN_CODE,
        "status": "active",
    }

    territory_state = {}
    for o in starting_ownerships:
        territory_id = o["territory_id"]
        owner_id = o["owner_id"]
        territory_state = set_territory_owner(
            territory_state, territory_id, owner_id
        )

    units = {}
    counters: Counters = {}
    territory_to_unit: TerritoryToUnit = {}
    for u in starting_units:
        units, territory_to_unit, counters = build_unit(
            units,
            territory_to_unit,
            counters,
            u["location_id"],
            u["unit_type"],
            u["owner_id"],
        )

    save(players, save_path / "players.json")
    save(units, save_path / "units.json")
    save(territory_state, save_path / "territory_state.json")
    save(game_meta, save_path / "game.json")
    save({}, save_path / "orders.json")

    gs = GameState(
        players=players,
        units=units,
        territory_state=territory_state,
        raw_orders={},
        game_meta=game_meta,
    )

    print(f"Game {game_id} created successfully!")
    return gs


def load_state(
    game_id: str, *, save_dir: str | os.PathLike[str] | None = None
) -> LoadedState:
    save_root = Path(save_dir or DEFAULT_GAMES_DIR)
    save_path = save_root / game_id

    gs = GameState(
        game_meta=load(save_path / "game.json"),
        players=load(save_path / "players.json"),
        territory_state=load(save_path / "territory_state.json"),
        units=load(save_path / "units.json"),
        raw_orders=load(save_path / "orders.json"),
    )

    territory_to_unit: TerritoryToUnit = build_territory_to_unit(gs.units)
    counters: Counters = build_counters(gs.units)
    dislodged: set[str] = gs.game_meta.get("dislodged", [])

    return LoadedState(
        game=gs,
        territory_to_unit=territory_to_unit,
        counters=counters,
        dislodged=dislodged,
    )


def build_territory_to_unit(
    units: dict[str, dict[str, Any]],
) -> TerritoryToUnit:
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
        "id": unit_id,
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
    territory_state[territory_id] = {
        "territory_id": territory_id,
        "owner_id": owner_id,
    }
    return territory_state


def eliminate_player(
    players: dict[str, dict[str, str]], player_id: str
) -> dict[str, Any]:
    players[player_id] = {"nation_id": player_id, "status": "eliminated"}
    return players
