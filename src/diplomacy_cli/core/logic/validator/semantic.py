from collections import deque

from ..rules_loader import Rules
from ..schema import (
    LoadedState,
    Order,
    OrderType,
    SemanticResult,
    SyntaxResult,
    TerritoryToUnit,
)


class SemanticError(Exception):
    pass


class InvalidSyntaxError(Exception):
    pass


def _check_territory_exists(prv: str, territory_ids: set) -> None:
    if prv not in territory_ids:
        raise SemanticError(f"{prv} is not a valid territory")


def _check_unit_exists(origin: str, territory_to_unit: TerritoryToUnit) -> None:
    if origin not in territory_to_unit:
        raise SemanticError(f"Unit does not exist in {origin}")


def _check_unit_ownership(
    player_id: str,
    origin: str,
    units: dict,
    territory_to_unit: TerritoryToUnit,
) -> None:
    unit_id = territory_to_unit[origin]
    owner_id = units[unit_id]["owner_id"]
    if player_id != owner_id:
        raise SemanticError(f"Unit in {origin} does not belong to {player_id}")


def _check_adjacency(
    origin: str,
    target: str,
    state: LoadedState,
    rules: Rules,
    allow_convoy: bool = False,
) -> None:
    _check_territory_exists(origin, rules.territory_ids)
    _check_territory_exists(target, rules.territory_ids)
    unit_id = state.territory_to_unit[origin]
    unit_type = state.game.units[unit_id]["unit_type"]

    for a, b, edge_type in rules.edges:
        if {a, b} == {origin, target}:
            if unit_type == "army" and edge_type in ("land", "both"):
                return
            if unit_type == "fleet" and edge_type in ("sea", "both"):
                return

    if allow_convoy and unit_type == "army":
        if _has_sea_path(origin, target, rules):
            return
        raise SemanticError(
            f"Army at {origin} cannot reach {target}: "
            "no continuous sea route for convoy"
        )

    raise SemanticError(
        f"{unit_type.title()} at {origin} cannot reach {target} "
        f"(requires {unit_type}-appropriate edge)"
    )


def _has_sea_path(origin: str, target: str, rules: Rules) -> bool:
    terr_type = rules.territory_type
    origin_coasts = rules.parent_to_coast.get(origin, {origin})
    target_coasts = rules.parent_to_coast.get(target, {target})

    visited = set(origin_coasts)
    queue = deque(origin_coasts)

    while queue:
        cur = queue.popleft()
        if cur in target_coasts:
            return True

        for a, b, mode in rules.edges:
            if mode not in ("sea", "both"):
                continue
            neigh = b if a == cur else a if b == cur else None
            if not neigh or neigh in visited:
                continue

            if terr_type.get(neigh) == "land" and neigh not in target_coasts:
                continue

            visited.add(neigh)
            queue.append(neigh)

    return False


def _check_hold(player_id: str, order: Order, state: LoadedState, rules: Rules):
    _check_territory_exists(order.origin, rules.territory_ids)
    _check_unit_exists(order.origin, state.territory_to_unit)
    _check_unit_ownership(
        player_id, order.origin, state.game.units, state.territory_to_unit
    )


def _check_support_hold(
    player_id: str, order: Order, state: LoadedState, rules: Rules
):
    assert order.support_origin is not None
    _check_territory_exists(order.origin, rules.territory_ids)
    _check_territory_exists(order.support_origin, rules.territory_ids)
    _check_unit_exists(order.origin, state.territory_to_unit)
    _check_unit_exists(order.support_origin, state.territory_to_unit)
    _check_adjacency(order.origin, order.support_origin, state, rules)
    _check_unit_ownership(
        player_id, order.origin, state.game.units, state.territory_to_unit
    )


def _check_convoy(
    player_id: str, order: Order, state: LoadedState, rules: Rules
):
    if not order.convoy_origin or not order.convoy_destination:
        raise SemanticError("Convoy must specify both origin and destination")

    for terr in (order.origin, order.convoy_origin, order.convoy_destination):
        _check_territory_exists(terr, rules.territory_ids)

    _check_unit_ownership(
        player_id, order.origin, state.game.units, state.territory_to_unit
    )
    _check_unit_exists(order.origin, state.territory_to_unit)
    fleet = state.game.units[state.territory_to_unit[order.origin]]
    if fleet["unit_type"] != "fleet":
        raise SemanticError(f"No fleet at {order.origin} to convoy")

    _check_unit_exists(order.convoy_origin, state.territory_to_unit)
    army = state.game.units[state.territory_to_unit[order.convoy_origin]]
    if army["unit_type"] != "army":
        raise SemanticError(f"No army at {order.convoy_origin} to convoy")

    if not _has_sea_path(order.convoy_origin, order.convoy_destination, rules):
        raise SemanticError(
            f"No valid sea path between {order.convoy_origin}"
            f" and {order.convoy_destination}"
        )


def _check_move(player_id: str, order: Order, state: LoadedState, rules: Rules):
    assert order.destination is not None
    _check_territory_exists(order.origin, rules.territory_ids)
    _check_territory_exists(order.destination, rules.territory_ids)
    _check_unit_ownership(
        player_id, order.origin, state.game.units, state.territory_to_unit
    )
    _check_adjacency(
        order.origin, order.destination, state, rules, allow_convoy=True
    )


def _check_support_move(
    player_id: str, order: Order, state: LoadedState, rules: Rules
):
    assert order.support_origin is not None
    assert order.support_destination is not None
    _check_territory_exists(order.origin, rules.territory_ids)
    _check_territory_exists(order.support_origin, rules.territory_ids)
    _check_territory_exists(order.support_destination, rules.territory_ids)
    _check_unit_exists(order.origin, state.territory_to_unit)
    _check_unit_exists(order.support_origin, state.territory_to_unit)
    _check_unit_ownership(
        player_id, order.origin, state.game.units, state.territory_to_unit
    )
    _check_adjacency(order.origin, order.support_destination, state, rules)
    _check_adjacency(
        order.support_origin, order.support_destination, state, rules
    )


def _check_build(
    player_id: str, order: Order, state: LoadedState, rules: Rules
) -> None:
    _check_territory_exists(order.origin, rules.territory_ids)
    prv = order.origin
    if prv not in rules.home_centers[player_id]:
        raise SemanticError(f"{prv} is not a home center of {player_id}")
    if state.game.territory_state[prv]["owner_id"] != player_id:
        raise SemanticError(f"{prv} does not belong to {player_id}")
    if order.origin in state.territory_to_unit:
        raise SemanticError(
            f"Cannot build in {order.origin}: territory is occupied"
        )
    count = sum(
        1
        for unit in state.game.units.values()
        if unit.get("owner_id") == player_id
    )
    supply_count = sum(
        1
        for prv in state.game.territory_state.values()
        if prv.get("owner_id") == player_id
    )
    if count >= supply_count:
        raise SemanticError(
            f"{player_id} does not have enough supply centers to build a unit"
        )
    if order.unit_type == "fleet" and order.origin not in rules.has_coast:
        raise SemanticError("Fleets can only be built on coasts")


def _check_disband(
    player_id: str, order: Order, state: LoadedState, rules: Rules
) -> None:
    _check_territory_exists(order.origin, rules.territory_ids)
    _check_unit_exists(order.origin, state.territory_to_unit)
    if (
        state.game.units[state.territory_to_unit[order.origin]]["unit_type"]
        != order.unit_type
    ):
        raise SemanticError(f"No {order.unit_type} at {order.origin}")
    _check_unit_ownership(
        player_id, order.origin, state.game.units, state.territory_to_unit
    )


def _check_retreat(
    player_id: str, order: Order, state: LoadedState, rules: Rules
) -> None:
    if order.destination is None:
        raise SemanticError("Retreat must specify a destination")

    if order.origin not in state.dislodged:
        raise SemanticError(f"No dislodged unit at {order.origin}")
    _check_territory_exists(order.origin, rules.territory_ids)
    _check_unit_exists(order.origin, state.territory_to_unit)
    _check_unit_ownership(
        player_id, order.origin, state.game.units, state.territory_to_unit
    )
    _check_territory_exists(order.destination, rules.territory_ids)
    _check_adjacency(order.origin, order.destination, state, rules)

    if order.destination in state.territory_to_unit:
        raise SemanticError(f"{order.destination} is occupied")


def validate_semantic(
    player_id: str, syntax: SyntaxResult, rules: Rules, state: LoadedState
):
    if not syntax.valid or syntax.order is None:
        raise InvalidSyntaxError(
            f"Cannot run semantic validation: syntax invalid ({syntax.errors})"
        )

    order = syntax.order
    errors: list[str] = []
    checker_map = {
        OrderType.MOVE: _check_move,
        OrderType.SUPPORT_MOVE: _check_support_move,
        OrderType.SUPPORT_HOLD: _check_support_hold,
        OrderType.CONVOY: _check_convoy,
        OrderType.BUILD: _check_build,
        OrderType.DISBAND: _check_disband,
        OrderType.HOLD: _check_hold,
        OrderType.RETREAT: _check_retreat,
    }

    checker = checker_map.get(order.order_type)
    if checker is None:
        raise SemanticError(f"Unhandled order type: {order.order_type}")

    try:
        checker(player_id, order, state, rules)
    except SemanticError as e:
        errors.append(str(e))

    return SemanticResult(
        raw=syntax.raw,
        normalized=syntax.normalized,
        order=order,
        valid=not errors,
        errors=errors,
    )
