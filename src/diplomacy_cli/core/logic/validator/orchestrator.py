from collections import defaultdict, Counter
from dataclasses import replace

from diplomacy_cli.core.logic.turn_code import Phase, parse_turn_code
from diplomacy_cli.core.logic.validator.resolution import (
    get_convoy_path,
    resolve_move_phase,
)
from diplomacy_cli.core.logic.validator.semantic import validate_semantic
from diplomacy_cli.core.logic.validator.syntax import parse_syntax

from ..schema import (
    LoadedState,
    Order,
    OrderType,
    OutcomeType,
    PhaseResolutionReport,
    ResolutionResult,
    Rules,
    SemanticResult,
)


def make_semantic_map(
    loaded_state: LoadedState,
    semantic_results: list[SemanticResult],
) -> tuple[dict[str, SemanticResult], dict[str, list[SemanticResult]]]:
    sem_by_unit = {}
    duplicate_sem_by_uid = defaultdict(list)
    for sem in semantic_results:
        uid = loaded_state.territory_to_unit[sem.order.origin]
        if uid in sem_by_unit:
            duplicate_sem_by_uid[uid].append(sem)
            continue
        sem_by_unit[uid] = sem

    for uid in loaded_state.game.units.keys():
        if uid not in sem_by_unit:
            hold_order = Order(
                origin=loaded_state.game.units[uid]["territory_id"],
                order_type=OrderType.HOLD,
            )
            sem_by_unit[uid] = SemanticResult(
                player_id=loaded_state.game.units[uid]["owner_id"],
                raw="",
                normalized="",
                order=hold_order,
                valid=True,
                errors=[],
            )
    return sem_by_unit, duplicate_sem_by_uid


def make_adjustment_semantic_map(
    loaded_state: LoadedState,
    semantic_results: list[SemanticResult],
) -> tuple[
    dict[str, SemanticResult],
    dict[str, list[SemanticResult]],
    dict[str, SemanticResult],
    dict[str, list[SemanticResult]],
]:
    disband_by_unit = {}
    duplicate_disband_by_uid = defaultdict(list)
    build_by_territory = {}
    duplicate_build_by_territory = defaultdict(list)
    for sem in semantic_results:
        if sem.order.order_type == OrderType.DISBAND:
            uid = loaded_state.territory_to_unit[sem.order.origin]
            if uid in disband_by_unit:
                duplicate_disband_by_uid[uid].append(sem)
                continue
            disband_by_unit[uid] = sem
        elif sem.order.order_type == OrderType.BUILD:
            territory = sem.order.origin
            if territory in build_by_territory:
                duplicate_build_by_territory[territory].append(sem)
                continue
            build_by_territory[territory] = sem

    return (
        disband_by_unit,
        duplicate_disband_by_uid,
        build_by_territory,
        duplicate_build_by_territory,
    )


def process_phase(
    raw_orders: dict[str, list[str]], rules: Rules, loaded_state: LoadedState
) -> PhaseResolutionReport:
    year, season, phase = parse_turn_code(
        loaded_state.game.game_meta["turn_code"]
    )
    validated_orders = []
    valid_syntax = []
    valid_semantics = []
    syntax_errors = []
    semantic_errors = []
    resolution_results = []
    for player, orders in raw_orders.items():
        for order in orders:
            parsed_order = parse_syntax(player, order, phase)
            if not parsed_order.valid:
                syntax_errors.append(parsed_order)
                continue
            valid_syntax.append(parsed_order)
            validated_order = validate_semantic(
                player, parsed_order, rules, loaded_state
            )
            if not validated_order.valid:
                semantic_errors.append(validated_order)
                continue
            valid_semantics.append(validated_order)
            validated_orders.append(validated_order)
    if phase in {Phase.MOVEMENT, Phase.RETREAT}:
        sem_by_unit, duplicated_orders_by_unit = make_semantic_map(
            loaded_state, validated_orders
        )
    match phase:
        case Phase.MOVEMENT:
            resolution_soa = resolve_move_phase(
                sem_by_unit, loaded_state, rules
            )
            n = len(resolution_soa.unit_id)
            for i in range(n):
                unit_id = resolution_soa.unit_id[i]
                origin_territory = resolution_soa.orig_territory[i]
                sem = sem_by_unit[unit_id]
                outcome = resolution_soa.outcome[i]
                dislodged_by_id = None
                supported_unit_id = None
                assert outcome is not None
                if outcome == OutcomeType.DISLODGED:
                    for j in range(n):
                        if resolution_soa.new_territory[j] == origin_territory:
                            dislodged_by_id = resolution_soa.unit_id[j]
                            break
                convoy_path = get_convoy_path(resolution_soa, i)
                if resolution_soa.order_type[i] in (
                    OrderType.SUPPORT_MOVE,
                    OrderType.SUPPORT_HOLD,
                ):
                    support_origin = resolution_soa.support_origin[i]
                    assert support_origin is not None
                    supported_unit_id = loaded_state.territory_to_unit[
                        support_origin
                    ]
                resolution_result = ResolutionResult(
                    unit_id=unit_id,
                    owner_id=resolution_soa.owner_id[i],
                    unit_type=resolution_soa.unit_type[i],
                    origin_territory=origin_territory,
                    semantic_result=sem,
                    outcome=outcome,
                    resolved_territory=resolution_soa.new_territory[i],
                    strength=resolution_soa.strength[i],
                    dislodged_by_id=dislodged_by_id,
                    destination=resolution_soa.move_destination[i],
                    convoy_path=convoy_path,
                    supported_unit_id=supported_unit_id,
                    duplicate_orders=duplicated_orders_by_unit.get(unit_id, []),
                )
                resolution_results.append(resolution_result)
        case Phase.RETREAT:
            retreat_results = []
            last_phase = loaded_state.pending_move
            if last_phase is None:
                raise ValueError(
                    "Retreat phase requires pending move data to process"
                )
            last_phase_resolution_results = last_phase.resolution_results
            occupied_territories = [
                r.resolved_territory
                for r in last_phase_resolution_results
                if r.outcome is not OutcomeType.DISLODGED
            ]
            for last_result in last_phase_resolution_results:
                if last_result.outcome == OutcomeType.DISLODGED:
                    assert last_result.unit_id is not None
                    sem = sem_by_unit[last_result.unit_id]
                    assert last_result.dislodged_by_id is not None
                    attacker_origin = loaded_state.game.units[
                        last_result.dislodged_by_id
                    ]["territory_id"]
                    assert sem.order.destination is not None
                    if (
                        sem.order.order_type == OrderType.HOLD
                        or sem.order.destination == attacker_origin
                        or sem.order.destination in occupied_territories
                    ):
                        outcome = OutcomeType.RETREAT_FAILED
                        resolved_territory = last_result.origin_territory
                    else:
                        outcome = OutcomeType.RETREAT_SUCCESS
                        resolved_territory = sem.order.destination
                    resolution_result = ResolutionResult(
                        unit_id=last_result.unit_id,
                        owner_id=last_result.owner_id,
                        unit_type=last_result.unit_type,
                        origin_territory=last_result.origin_territory,
                        semantic_result=sem,
                        outcome=outcome,
                        resolved_territory=resolved_territory,
                        strength=1,
                        dislodged_by_id=last_result.dislodged_by_id,
                        destination=sem.order.destination,
                        duplicate_orders=duplicated_orders_by_unit.get(
                            last_result.unit_id, []
                        ),
                    )
                    retreat_results.append(resolution_result)
            destination_counts = Counter(
                r.resolved_territory
                for r in retreat_results
                if r.outcome == OutcomeType.RETREAT_SUCCESS
            )
            resolution_results = [
                replace(
                    r,
                    outcome=OutcomeType.RETREAT_FAILED,
                    resolved_territory=r.origin_territory,
                )
                if r.outcome == OutcomeType.RETREAT_SUCCESS
                and destination_counts[r.resolved_territory] > 1
                else r
                for r in retreat_results
            ]
        case Phase.ADJUSTMENT:
            (
                disband_by_id,
                duplicate_disband_by_id,
                build_by_territory,
                duplicate_build_by_territory,
            ) = make_adjustment_semantic_map(loaded_state, validated_orders)
            unit_count = {}
            supply_center_count = {}
            resolution_results = []
            for player in loaded_state.game.players.values():
                nation = player["nation_id"]
                unit_count[nation] = 0
                for unit in loaded_state.game.units.values():
                    if unit["owner_id"] == nation:
                        unit_count[nation] += 1
                supply_center_count[nation] = sum(
                    1
                    for prv in loaded_state.game.territory_state.values()
                    if prv.get("owner_id") == nation
                )

            for unit_id, disband in disband_by_id.items():
                unit_count[disband.player_id] -= 1
                disband_result = ResolutionResult(
                    unit_id=unit_id,
                    owner_id=disband.player_id,
                    unit_type=loaded_state.game.units[unit_id]["unit_type"],
                    origin_territory=disband.order.origin,
                    semantic_result=disband,
                    outcome=OutcomeType.DISBAND_SUCCESS,
                    resolved_territory=disband.order.origin,
                    strength=1,
                    duplicate_orders=duplicate_disband_by_id.get(unit_id, []),
                )
                resolution_results.append(disband_result)

            for territory, build in build_by_territory.items():
                if (
                    unit_count[build.player_id] + 1
                    > supply_center_count[build.player_id]
                ):
                    outcome = OutcomeType.BUILD_NO_CENTER
                else:
                    unit_count[build.player_id] += 1
                    outcome = OutcomeType.BUILD_SUCCESS
                assert build.order.unit_type is not None
                build_result = ResolutionResult(
                    unit_id=None,
                    owner_id=build.player_id,
                    unit_type=build.order.unit_type,
                    origin_territory=territory,
                    semantic_result=build,
                    outcome=outcome,
                    resolved_territory=territory,
                    strength=1,
                    duplicate_orders=duplicate_build_by_territory.get(
                        territory, []
                    ),
                )
                resolution_results.append(build_result)
    return PhaseResolutionReport(
        phase=phase,
        season=season,
        year=year,
        valid_syntax=valid_syntax,
        valid_semantics=valid_semantics,
        syntax_errors=syntax_errors,
        semantic_errors=semantic_errors,
        resolution_results=resolution_results,
    )
