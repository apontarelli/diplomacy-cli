from collections import defaultdict
from dataclasses import dataclass

from diplomacy_cli.core.logic.schema import (
    LoadedState,
    Order,
    OrderType,
    ResolutionSoA,
    SemanticResult,
    OutcomeType,
)


@dataclass(frozen=True)
class ResolutionMaps:
    move_by_origin: dict[str, int]
    moves_by_dest: dict[str, list[int]]
    support_by_origin: dict[str, int]
    support_moves_by_supported_origin: dict[str, list[int]]
    support_moves_by_supported_dest: dict[str, list[int]]
    support_holds_by_supported_origin: dict[str, list[int]]
    convoy_by_origin: dict[str, int]
    convoys_by_army_origin: dict[str, list[int]]
    convoys_by_army_dest: dict[str, list[int]]
    hold_by_origin: dict[str, int]


def make_resolution_maps(soa: ResolutionSoA) -> ResolutionMaps:
    move_by_origin = {}
    moves_by_dest = defaultdict(list)
    support_by_origin = {}
    support_moves_by_supported_origin = defaultdict(list)
    support_moves_by_supported_dest = defaultdict(list)
    support_holds_by_supported_origin = defaultdict(list)
    convoy_by_origin = {}
    convoys_by_army_origin = defaultdict(list)
    convoys_by_army_dest = defaultdict(list)
    hold_by_origin = {}
    for idx, ot in enumerate(soa.order_type):
        match ot:
            case OrderType.MOVE:
                move_by_origin[soa.orig_territory[idx]] = idx
                moves_by_dest[soa.move_destination[idx]].append(idx)
            case OrderType.SUPPORT_MOVE:
                support_by_origin[soa.orig_territory[idx]] = idx
                support_moves_by_supported_dest[
                    soa.support_destination[idx]
                ].append(idx)
                support_moves_by_supported_origin[
                    soa.support_origin[idx]
                ].append(idx)
            case OrderType.SUPPORT_HOLD:
                support_by_origin[soa.orig_territory[idx]] = idx
                support_holds_by_supported_origin[
                    soa.support_origin[idx]
                ].append(idx)
            case OrderType.HOLD:
                hold_by_origin[soa.orig_territory[idx]] = idx
            case OrderType.CONVOY:
                convoy_by_origin[soa.orig_territory[idx]] = idx
                convoys_by_army_origin[soa.convoy_origin[idx]].append(idx)
                convoys_by_army_dest[soa.convoy_destination[idx]].append(idx)

    return ResolutionMaps(
        move_by_origin=move_by_origin,
        moves_by_dest=moves_by_dest,
        support_by_origin=support_by_origin,
        support_moves_by_supported_origin=support_moves_by_supported_origin,
        support_moves_by_supported_dest=support_moves_by_supported_dest,
        support_holds_by_supported_origin=support_holds_by_supported_origin,
        convoy_by_origin=convoy_by_origin,
        convoys_by_army_origin=convoys_by_army_origin,
        convoys_by_army_dest=convoys_by_army_dest,
        hold_by_origin=hold_by_origin,
    )


def make_semantic_map(
    loaded_state: LoadedState,
    semantic_results: list[SemanticResult],
) -> tuple[dict[str, SemanticResult], list[str]]:
    sem_by_unit: dict[str, SemanticResult] = {}
    duplicate_orders: list[str] = []
    for sem in semantic_results:
        uid = loaded_state.territory_to_unit[sem.order.origin]
        if uid in sem_by_unit:
            duplicate_orders.append(
                f"Duplicate order for unit {uid}; ignoring."
            )
            continue
        sem_by_unit[uid] = sem

    for uid in loaded_state.game.units.keys():
        if uid not in sem_by_unit:
            hold_order = Order(
                origin=loaded_state.game.units[uid]["territory_id"],
                order_type=OrderType.HOLD,
            )
            sem_by_unit[uid] = SemanticResult(
                raw="", normalized="", order=hold_order, valid=True, errors=[]
            )
    return sem_by_unit, duplicate_orders


def move_phase_soa(
    loaded_state: LoadedState, semantic_results: list[SemanticResult]
) -> tuple[ResolutionSoA, list[str]]:
    sem_by_unit, duplicate_orders = make_semantic_map(
        loaded_state, semantic_results
    )
    unit_id = []
    owner_id = []
    unit_type = []
    orig_territory = []
    order_type = []
    move_destination = []
    support_origin = []
    support_destination = []
    convoy_origin = []
    convoy_destination = []

    for uid, u_data in loaded_state.game.units.items():
        sem = sem_by_unit[uid]

        unit_id.append(uid)
        owner_id.append(u_data["owner_id"])
        unit_type.append(u_data["unit_type"])
        orig_territory.append(u_data["territory_id"])

        o = sem.order
        order_type.append(o.order_type)
        move_destination.append(o.destination)
        support_origin.append(o.support_origin)
        support_destination.append(o.support_destination)
        convoy_origin.append(o.convoy_origin)
        convoy_destination.append(o.convoy_destination)

    n = len(unit_id)
    soa = ResolutionSoA(
        unit_id=unit_id,
        owner_id=owner_id,
        unit_type=unit_type,
        orig_territory=orig_territory,
        order_type=order_type,
        move_destination=move_destination,
        support_origin=support_origin,
        support_destination=support_destination,
        convoy_origin=convoy_origin,
        convoy_destination=convoy_destination,
        new_territory=orig_territory.copy(),
        strength=[1] * n,
        dislodged=[False] * n,
        support_cut=[False] * n,
        has_valid_convoy=[False] * n,
        is_resolved=[False] * n,
        outcome=[None] * n,
    )

    return soa, duplicate_orders


def validate_convoy(
    soa: ResolutionSoA, maps: ResolutionMaps, idx: int
) -> OutcomeType | None:
    orig = soa.convoy_origin[idx]
    if orig is None:
        return OutcomeType.INVALID_CONVOY
    move_idx = maps.move_by_origin.get(orig)
    if (
        move_idx is None
        or soa.move_destination[move_idx] != soa.convoy_destination[idx]
    ):
        return OutcomeType.INVALID_CONVOY
    return None


def validate_support_move(
    soa: ResolutionSoA, maps: ResolutionMaps, idx: int
) -> OutcomeType | None:
    orig = soa.support_origin[idx]
    if orig is None:
        return OutcomeType.INVALID_SUPPORT
    move_idx = maps.move_by_origin.get(orig)
    if (
        move_idx is None
        or soa.move_destination[move_idx] != soa.support_destination[idx]
    ):
        return OutcomeType.INVALID_SUPPORT
    return None


def validate_support_hold(
    soa: ResolutionSoA, maps: ResolutionMaps, idx: int
) -> OutcomeType | None:
    orig = soa.support_origin[idx]
    if orig is None:
        return OutcomeType.INVALID_SUPPORT
    hold_idx = maps.hold_by_origin.get(orig)
    if hold_idx is None or soa.order_type[hold_idx] != OrderType.HOLD:
        return OutcomeType.INVALID_SUPPORT
    return None


def flag_support_convoy_mismatches(
    soa: ResolutionSoA, maps: ResolutionMaps
) -> list[OutcomeType | None]:
    result: list[OutcomeType | None] = [None] * len(soa.order_type)
    for idx, typ in enumerate(soa.order_type):
        match typ:
            case OrderType.CONVOY:
                result[idx] = validate_convoy(soa, maps, idx)
            case OrderType.SUPPORT_MOVE:
                result[idx] = validate_support_move(soa, maps, idx)
            case OrderType.SUPPORT_HOLD:
                result[idx] = validate_support_hold(soa, maps, idx)
    return result
