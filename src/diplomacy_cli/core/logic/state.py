from __future__ import annotations

from collections import defaultdict
import json
from pathlib import Path
from typing import Any

from diplomacy_cli.core.logic.rules_loader import load_rules
from diplomacy_cli.core.logic.validator.orchestrator import process_phase

from diplomacy_cli.core.logic.schema import (
    Counters,
    GameState,
    LoadedState,
    OutcomeType,
    Phase,
    PhaseResolutionReport,
    Season,
    TerritoryToUnit,
    UnitType,
)
from diplomacy_cli.core.logic.storage import (
    load,
    load_variant_json,
    save,
)
from diplomacy_cli.core.logic.turn_code import (
    INITIAL_TURN_CODE,
    advance_turn_code,
    parse_turn_code,
)
from diplomacy_cli.core.paths import (
    DEFAULT_GAMES_DIR,
    GamePaths,
    ensure_dir,
    game_dir,
    game_meta_path,
    orders_path,
    players_path,
    report_path,
    territory_state_path,
    units_path,
)
from .serialization import (
    phase_resolution_report_from_dict,
    phase_resolution_report_to_dict,
)


def start_game(
    game_id: str = "new_game",
    *,
    variant: str = "classic",
    root_dir: Path = DEFAULT_GAMES_DIR,
) -> GameState:
    base = Path(root_dir)
    paths = GamePaths(base, game_id)
    gd = game_dir(paths)

    if gd.exists():
        raise FileExistsError(f"Save directory '{gd}' already exists.")
    ensure_dir(gd)

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

    save(players, players_path(paths))
    save(units, units_path(paths))
    save(territory_state, territory_state_path(paths))
    save(game_meta, game_meta_path(paths))
    save({}, orders_path(paths))

    gs = GameState(
        players=players,
        units=units,
        territory_state=territory_state,
        raw_orders={},
        game_meta=game_meta,
    )

    print(f"Game {game_id} created successfully!")
    return gs


def load_state(game_id: str, root_dir: Path = DEFAULT_GAMES_DIR) -> LoadedState:
    base = Path(root_dir)
    paths = GamePaths(base, game_id)
    gs = GameState(
        game_meta=load(game_meta_path(paths)),
        players=load(players_path(paths)),
        territory_state=load(territory_state_path(paths)),
        units=load(units_path(paths)),
        raw_orders=load(orders_path(paths)),
    )

    for udata in gs.units.values():
        ut = udata.get("unit_type")
        if isinstance(ut, str):
            udata["unit_type"] = UnitType(ut)

    territory_to_unit: TerritoryToUnit = build_territory_to_unit(gs.units)
    counters: Counters = build_counters(gs.units)
    year, season, phase = parse_turn_code(gs.game_meta["turn_code"])
    pending_move = None
    if phase == Phase.RETREAT:
        pending_move = load_phase_resolution_report(
            game_id, year, season, Phase.RETREAT
        )

    return LoadedState(
        game=gs,
        territory_to_unit=territory_to_unit,
        counters=counters,
        pending_move=pending_move,
    )


def load_orders() -> dict[str, list[str]]:
    raw_orders = defaultdict(list)
    return raw_orders


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


def save_phase_resolution_report(
    game_id: str,
    phase_resolution_report: PhaseResolutionReport,
    root_dir: Path = DEFAULT_GAMES_DIR,
) -> None:
    base = Path(root_dir)
    paths = GamePaths(base, game_id)

    phase_report_dict = phase_resolution_report_to_dict(phase_resolution_report)
    save(
        phase_report_dict,
        report_path(
            paths,
            phase_resolution_report.year,
            phase_resolution_report.season,
            phase_resolution_report.phase,
        ),
    )


def load_phase_resolution_report(
    game_id: str,
    year: int,
    season: Season,
    phase: Phase,
    root_dir: Path = DEFAULT_GAMES_DIR,
) -> PhaseResolutionReport:
    base = Path(root_dir)
    paths = GamePaths(base, game_id)

    phase_report_dict = load(
        report_path(
            paths,
            year,
            season,
            phase,
        ),
    )

    return phase_resolution_report_from_dict(phase_report_dict)


def apply_state_mutations(
    loaded_state: LoadedState, phase_resolution_report: PhaseResolutionReport
) -> LoadedState:
    game_state = loaded_state.game
    units = game_state.units
    territory_to_unit = loaded_state.territory_to_unit
    counters = loaded_state.counters
    territory_state = game_state.territory_state
    if phase_resolution_report.phase == Phase.RETREAT:
        movements = []
        for result in phase_resolution_report.resolution_results:
            if result.outcome == OutcomeType.RETREAT_FAILED:
                units, territory_to_unit = disband_unit(
                    units,
                    territory_to_unit,
                    result.origin_territory,
                )
            if result.outcome in (
                OutcomeType.RETREAT_SUCCESS,
                OutcomeType.MOVE_SUCCESS,
            ):
                movements.append(
                    {"from": result.origin_territory, "to": result.destination}
                )
        units, territory_to_unit = apply_unit_movements(
            units, territory_to_unit, movements
        )
    elif phase_resolution_report.phase == Phase.ADJUSTMENT:
        for result in phase_resolution_report.resolution_results:
            if result.outcome == OutcomeType.DISBAND_SUCCESS:
                units, territory_to_unit = disband_unit(
                    units,
                    territory_to_unit,
                    result.origin_territory,
                )
            if result.outcome == OutcomeType.BUILD_SUCCESS:
                units, territory_to_unit, counters = build_unit(
                    units,
                    territory_to_unit,
                    counters,
                    result.origin_territory,
                    result.unit_type,
                    result.owner_id,
                )
        for territory_id in territory_state:
            if territory_id in territory_to_unit:
                uid = territory_to_unit[territory_id]
                unit = units[uid]
                if (
                    unit["owner_id"]
                    != territory_state[territory_id]["owner_id"]
                ):
                    territory_state = set_territory_owner(
                        territory_state, territory_id, unit["owner_id"]
                    )

    game_state = GameState(
        game_state.players,
        units,
        territory_state,
        game_state.raw_orders,
        game_state.game_meta,
    )
    return LoadedState(game_state, territory_to_unit, counters, None)


def process_turn(
    game_id: str,
    root_dir: Path = DEFAULT_GAMES_DIR,
) -> LoadedState:
    base = Path(root_dir)
    paths = GamePaths(base, game_id)

    loaded_state = load_state(game_id, base)
    rules = load_rules(loaded_state.game.game_meta["variant"])
    report = process_phase(loaded_state, rules)

    current_turn = loaded_state.game.game_meta["turn_code"]
    dislodged = any(
        r.outcome == OutcomeType.DISLODGED for r in report.resolution_results
    )

    if report.phase == Phase.MOVEMENT:
        save_phase_resolution_report(game_id, report, base)
        if dislodged:
            new_turn = advance_turn_code(current_turn, False)
        else:
            loaded_state = apply_state_mutations(loaded_state, report)
            new_turn = advance_turn_code(current_turn, True)
    elif report.phase in (Phase.RETREAT, Phase.ADJUSTMENT):
        loaded_state = apply_state_mutations(loaded_state, report)
        save_phase_resolution_report(game_id, report, base)
        new_turn = advance_turn_code(current_turn, False)

    game_meta = dict(loaded_state.game.game_meta)
    game_meta["turn_code"] = new_turn

    updated_players = {}
    for pid, pdata in loaded_state.game.players.items():
        owns_territory = any(
            t["owner_id"] == pid
            for t in loaded_state.game.territory_state.values()
        )
        if not owns_territory:
            pdata = {**pdata, "status": "eliminated"}
        updated_players[pid] = pdata

    updated_state = GameState(
        players=updated_players,
        units=loaded_state.game.units,
        territory_state=loaded_state.game.territory_state,
        raw_orders={},
        game_meta=game_meta,
    )

    final_state = LoadedState(
        updated_state,
        build_territory_to_unit(updated_state.units),
        build_counters(updated_state.units),
    )

    save(updated_state.players, players_path(paths))
    save(updated_state.units, units_path(paths))
    save(updated_state.territory_state, territory_state_path(paths))
    save(updated_state.game_meta, game_meta_path(paths))
    save({}, orders_path(paths))

    return final_state


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
    key = f"{owner_id}_{unit_type.lower()}"
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
