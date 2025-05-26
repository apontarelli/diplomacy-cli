from collections import defaultdict, deque
from dataclasses import dataclass, replace

from diplomacy_cli.core.logic.schema import (
    LoadedState,
    Order,
    OrderType,
    OutcomeType,
    ResolutionSoA,
    Rules,
    SemanticResult,
    UnitType,
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
                hold_by_origin[soa.orig_territory[idx]] = idx
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
        convoy_path_flat=[],
        convoy_path_len=[0] * n,
        convoy_path_start=[-1] * n,
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


def find_convoy_path(
    origin: str,
    destination: str,
    convoy_fleet_territories: list[str],
    rules: Rules,
) -> list[str] | None:
    origin_coasts = rules.parent_to_coast.get(origin, {origin})
    destination_coasts = rules.parent_to_coast.get(destination, {destination})

    visited = set(origin_coasts)
    queue = deque(origin_coasts)
    previous: dict[str, str] = {}

    while queue:
        cur = queue.popleft()
        if cur in destination_coasts:
            path = [cur]
            while cur in previous:
                cur = previous[cur]
                path.append(cur)
            return list(reversed(path))

        for neighbor, mode in rules.adjacency_map.get(cur, []):
            if mode not in ("sea", "both"):
                continue
            if (
                neighbor not in convoy_fleet_territories
                and neighbor not in destination_coasts
            ):
                continue
            if neighbor in visited:
                continue

            visited.add(neighbor)
            previous[neighbor] = cur
            queue.append(neighbor)

    return None


def get_convoy_path(soa: ResolutionSoA, idx: int) -> list[str] | None:
    start = soa.convoy_path_start[idx]
    if start == -1:
        return None
    length = soa.convoy_path_len[idx]
    return soa.convoy_path_flat[start : start + length]


def process_convoys(
    soa: ResolutionSoA,
    rules: Rules,
    origin_to_move: dict,
    origin_to_convoy: dict,
) -> tuple[list[int], list[int], list[str]]:
    n = len(soa.order_type)
    path_start = [-1] * n
    path_len = [0] * n
    path_flat = []
    convoy_support_map: dict[int, list[str]] = defaultdict(list)
    for idx in origin_to_convoy.values():
        if soa.dislodged[idx] or soa.outcome[idx] == OutcomeType.INVALID_CONVOY:
            continue
        move_idx = origin_to_move[soa.convoy_origin[idx]]
        convoy_fleet_origin = soa.orig_territory[idx]
        convoy_support_map[move_idx].append(convoy_fleet_origin)

    for move_idx, convoy_list in convoy_support_map.items():
        move_origin = soa.orig_territory[move_idx]
        move_destination = soa.move_destination[move_idx]
        assert move_destination is not None, (
            f"Expected move destination at index {move_idx}"
        )
        path = find_convoy_path(
            move_origin, move_destination, convoy_list, rules
        )
        if path is None:
            path_start[move_idx] = -1
            path_len[move_idx] = 0
        else:
            start = len(path_flat)
            path_start[move_idx] = start
            path_len[move_idx] = len(path)
            path_flat.extend(path)

    return path_start, path_len, path_flat


def process_moves(
    soa: ResolutionSoA, move_by_origin: dict[str, int], rules: Rules
) -> tuple[list[str], list[OutcomeType | None]]:
    new_territory: list[str] = soa.new_territory.copy()
    outcome: list[OutcomeType | None] = soa.outcome.copy()
    for origin, move_idx in move_by_origin.items():
        destination = soa.move_destination[move_idx]
        assert destination is not None
        unit_type = soa.unit_type[move_idx]
        is_adjacent = any(
            adj == destination for adj, _ in rules.adjacency_map[origin]
        )

        if unit_type == UnitType.ARMY:
            has_convoy = soa.convoy_path_len[move_idx] > 0
            if has_convoy or is_adjacent:
                new_territory[move_idx] = destination
            else:
                outcome[move_idx] = OutcomeType.MOVE_NO_CONVOY
        elif unit_type == UnitType.FLEET:
            if is_adjacent:
                new_territory[move_idx] = destination
    return new_territory, outcome


def cut_supports(
    soa: ResolutionSoA, move_by_origin: dict[str, int]
) -> list[bool]:
    n = len(soa.order_type)
    support_cut = [False] * n
    for i in range(n):
        order_type = soa.order_type[i]
        if order_type not in (OrderType.SUPPORT_MOVE, OrderType.SUPPORT_HOLD):
            continue

        supporter_territory = soa.orig_territory[i]
        support_target = soa.support_destination[i]

        for move_origin, move_idx in move_by_origin.items():
            if soa.new_territory[move_idx] != supporter_territory:
                continue
            if soa.outcome[move_idx] == OutcomeType.MOVE_NO_CONVOY:
                continue
            if move_origin == support_target:
                continue

            support_cut[i] = True
            break

    return support_cut


def calculate_strength(soa: ResolutionSoA, maps: ResolutionMaps):
    strength = [1] * len(soa.order_type)
    for (
        supported_origin,
        support_idxs,
    ) in maps.support_moves_by_supported_origin.items():
        for support_idx in support_idxs:
            if soa.support_cut[support_idx]:
                continue
            supported_idx = maps.move_by_origin[supported_origin]
            strength[supported_idx] += 1
    for (
        supported_origin,
        support_idxs,
    ) in maps.support_holds_by_supported_origin.items():
        for support_idx in support_idxs:
            if soa.support_cut[support_idx]:
                continue
            supported_idx = maps.hold_by_origin[supported_origin]
            strength[supported_idx] += 1
    return strength


def resolve_conflict(soa: ResolutionSoA) -> list[str]:
    new_territory = soa.new_territory.copy()
    strength = soa.strength

    changed = True
    while changed:
        changed = False
        contested = defaultdict(list)

        for i, dest in enumerate(new_territory):
            contested[dest].append(i)

        for indices in contested.values():
            if len(indices) == 1:
                continue
            s = [strength[i] for i in indices]
            max_s = max(s)
            winners = [i for i in indices if strength[i] == max_s]

            if len(winners) > 1:
                for i in indices:
                    if new_territory[i] != soa.orig_territory[i]:
                        new_territory[i] = soa.orig_territory[i]
                        changed = True
            else:
                winner = winners[0]
                for i in indices:
                    if (
                        i != winner
                        and new_territory[i] != soa.orig_territory[i]
                    ):
                        new_territory[i] = soa.orig_territory[i]
                        changed = True

    return new_territory


def detect_dislodged(soa: ResolutionSoA) -> list[bool]:
    n = len(soa.unit_type)
    dislodged = [False] * n
    for i in range(n):
        if soa.new_territory[i] == soa.orig_territory[i]:
            for j in range(n):
                if j == i:
                    continue
                if soa.new_territory[j] == soa.orig_territory[i]:
                    if soa.owner_id[j] != soa.owner_id[i]:
                        dislodged[i] = True
                        break
    return dislodged


def assign_move_outcomes(soa: ResolutionSoA) -> list[OutcomeType | None]:
    outcome = soa.outcome.copy()
    n = len(soa.unit_id)

    for i in range(n):
        if outcome[i] is not None:
            continue
        if soa.dislodged[i]:
            outcome[i] = OutcomeType.DISLODGED
            continue

        match soa.order_type[i]:
            case OrderType.MOVE:
                if soa.move_destination[i] == soa.new_territory[i]:
                    outcome[i] = OutcomeType.MOVE_SUCCESS
                elif soa.orig_territory[i] == soa.new_territory[i]:
                    outcome[i] = OutcomeType.MOVE_BOUNCED
                else:
                    raise ValueError(
                        f"Inconsistent new_territory for MOVE at index {i}"
                    )
            case OrderType.SUPPORT_HOLD | OrderType.SUPPORT_MOVE:
                if soa.support_cut[i]:
                    outcome[i] = OutcomeType.SUPPORT_CUT
                else:
                    outcome[i] = OutcomeType.SUPPORT_SUCCESS
            case OrderType.HOLD:
                outcome[i] = OutcomeType.HOLD_SUCCESS
            case OrderType.CONVOY:
                outcome[i] = OutcomeType.CONVOY_SUCCESS

    return outcome
