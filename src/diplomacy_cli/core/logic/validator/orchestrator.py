from collections import defaultdict

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


def process_move_phase(
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
            parsed_order = parse_syntax(player, order, Phase.MOVEMENT)
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
    sem_by_unit, duplicated_orders_by_unit = make_semantic_map(
        loaded_state, validated_orders
    )
    resolution_soa = resolve_move_phase(sem_by_unit, loaded_state, rules)
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
            supported_unit_id = loaded_state.territory_to_unit[support_origin]
        resolution_result = ResolutionResult(
            unit_id=resolution_soa.unit_id[i],
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
