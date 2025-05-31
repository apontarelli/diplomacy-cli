from collections import defaultdict

from ..schema import (
    LoadedState,
    Order,
    OrderType,
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
