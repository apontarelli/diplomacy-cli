from collections import defaultdict
from typing import List

from diplomacy_cli.core.logic.schema import (
    LoadedState,
    OutcomeType,
    PhaseResolutionReport,
    Rules,
)
from diplomacy_cli.core.logic.turn_code import BASE_YEAR


def format_state(loaded_state: LoadedState):
    state = loaded_state.game
    game = format_game(state.game_meta)
    players = format_players(state.players)
    territory_state = format_territory_state(state.territory_state)
    units = format_units(state.units)

    output = []
    output.append(f"\n{game}")
    output.append(f"\n{players}")
    output.append(f"\n{territory_state}")
    output.append(f"\n{units}")

    return "\n".join(output)


def format_orders(player_orders: list[str], player: str):
    output = []
    output.append(f"--- Orders: {player} ---")
    if player_orders == []:
        output.append(f"No orders found")
        return "\n".join(output)
    i = 1
    for order in player_orders:
        output.append(f"{i}: {order}")
        i += 1
    return "\n".join(output)


def format_game(game):
    output = []
    output.append(f"--- Game: {game['game_id']} ---")
    output.append(f"Variant: {game['variant']}")
    output.append(f"Turn: {game['turn_code']}")
    output.append(f"Status: {game['status']}")

    return "\n".join(output)


def format_phase_resolution_report(
    phase_report: PhaseResolutionReport, rules: Rules
) -> str:
    output: List[str] = []

    season_name = phase_report.season.name.capitalize()
    phase_name = phase_report.phase.name.capitalize()
    territory_names = rules.territory_display_names
    nation_names = rules.nation_display_names

    output.append(
        f"=== {phase_report.year + BASE_YEAR} {season_name} - {phase_name} Phase ==="
    )
    output.append("")

    output.append("-- Syntax Validation --")
    if phase_report.valid_syntax:
        output.append("Valid syntax orders:")
        for s in phase_report.valid_syntax:
            output.append(f"  [{s.player_id}] {s.normalized}")
    else:
        output.append("No valid syntax orders.")
    if phase_report.syntax_errors:
        output.append("Syntax errors:")
        for s in phase_report.syntax_errors:
            errs = "; ".join(s.errors)
            output.append(f"  [{s.player_id}] {s.raw} -> {errs}")
    output.append("")

    output.append("-- Semantic Validation --")
    if phase_report.valid_semantics:
        output.append("Valid semantic orders:")
        for sem in phase_report.valid_semantics:
            output.append(f"  [{sem.player_id}] {sem.normalized}")
    else:
        output.append("No valid semantic orders.")
    if phase_report.semantic_errors:
        output.append("Semantic errors:")
        for sem in phase_report.semantic_errors:
            errs = "; ".join(sem.errors)
            output.append(f"  [{sem.player_id}] {sem.normalized} -> {errs}")
    output.append("")

    output.append("-- Resolution Results --")
    if not phase_report.resolution_results:
        output.append("No resolution results.")
    else:
        for res in phase_report.resolution_results:
            base = f"{nation_names[res.owner_id]} ({res.unit_type.title()}) at {territory_names[res.origin_territory]}"
            outcome = res.outcome
            if outcome == OutcomeType.MOVE_SUCCESS:
                base += f" -> moved to {res.resolved_territory} (strength {res.strength})"
            elif outcome == OutcomeType.MOVE_BOUNCED:
                base += f" -> move bounced at {res.destination or res.resolved_territory}"
            elif outcome == OutcomeType.MOVE_NO_CONVOY:
                base += f" -> move failed (no convoy) to {res.destination}"
            elif outcome == OutcomeType.SUPPORT_SUCCESS:
                base += f" -> support succeeded"
            elif outcome == OutcomeType.SUPPORT_CUT:
                base += f" -> support cut"
            elif outcome == OutcomeType.INVALID_SUPPORT:
                base += f" -> invalid support"
            elif outcome == OutcomeType.HOLD_SUCCESS:
                base += f" -> held"
            elif outcome == OutcomeType.CONVOY_SUCCESS:
                path = "->".join(res.convoy_path or [])
                base += f" -> convoyed to {res.resolved_territory} via {path}"
            elif outcome == OutcomeType.INVALID_CONVOY:
                base += f" -> invalid convoy"
            elif outcome == OutcomeType.DISLODGED:
                base += f" -> dislodged by {res.dislodged_by_id}"
            elif outcome == OutcomeType.RETREAT_SUCCESS:
                base += f" -> retreated to {res.resolved_territory}"
            elif outcome == OutcomeType.RETREAT_FAILED:
                base += f" -> failed to retreat"
            elif outcome == OutcomeType.BUILD_SUCCESS:
                base += f" -> built at {res.resolved_territory}"
            elif outcome == OutcomeType.BUILD_ILLEGAL_LOCATION:
                base += f" -> build illegal location"
            elif outcome == OutcomeType.BUILD_NO_CENTER:
                base += f" -> build failed (no center)"
            elif outcome == OutcomeType.DISBAND_SUCCESS:
                base += f" -> disbanded"
            elif outcome == OutcomeType.DISBAND_FAILED:
                base += f" -> failed to disband"
            else:
                base += f" -> outcome: {outcome.value}"

            output.append(base)

            if res.duplicate_orders:
                dup_str = "; ".join(d.normalized for d in res.duplicate_orders)
                output.append(f"    Duplicate orders: {dup_str}")

            if res.convoy_path and outcome != OutcomeType.CONVOY_SUCCESS:
                path = "->".join(res.convoy_path)
                output.append(f"    Convoy path attempted: {path}")
    return "\n".join(output)


def format_players(players, status: str = "both"):
    output = []
    match status:
        case "both":
            output.append("--- Players ---")
            for k, v in players.items():
                output.append(f"{k} - {v['status']}")
        case "active":
            output.append("--- Active Players ---")
            for k, v in players.items():
                if v["status"] == status:
                    output.append(f"{k} - {v['status']}")
        case "eliminated":
            output.append("--- Active Players ---")
            for k, v in players.items():
                if v["status"] == status:
                    output.append(f"{k} - {v['status']}")

    return "\n".join(output)


def format_territory_state(territory_state):
    output = []
    by_owner = group_territory_state_by_owner(territory_state)

    output.append("--- Territories ---")
    for owner, owned_territories in by_owner.items():
        output.append(f"\n{owner}")
        for territory_id in owned_territories:
            output.append(f" - {territory_id}")

    return "\n".join(output)


def format_units(units):
    output = []
    by_owner = group_units_by_owner(flatten_units(units))

    output.append("--- Units ---")
    for owner, units in by_owner.items():
        output.append(f"\n{owner}")
        for u in units:
            output.append(f" - {u['unit_type'].lower()} - {u['territory_id']}")

    return "\n".join(output)


def flatten_units(units_dict):
    return [
        {"territory_id": territory, **unit}
        for territory, unit in units_dict.items()
    ]


def group_territory_state_by_owner(territory_state_dict):
    grouped = defaultdict(list)
    for territory_id, data in territory_state_dict.items():
        owner = data.get("owner_id", "None")
        grouped[owner].append(territory_id)
    return grouped


def group_units_by_owner(unit_list):
    grouped = defaultdict(list)
    for unit in unit_list:
        owner = unit.get("owner_id", "None")
        grouped[owner].append(unit)
    return grouped
