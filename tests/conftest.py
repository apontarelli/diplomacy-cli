from types import SimpleNamespace
from typing import cast

import pytest

from diplomacy_cli.core.logic.rules_loader import load_rules
from diplomacy_cli.core.logic.schema import (
    GameState,
    LoadedState,
    Order,
    OrderType,
    OutcomeType,
    ResolutionSoA,
    Rules,
    SemanticResult,
    UnitType,
)
from diplomacy_cli.core.logic.state import (
    build_counters,
    build_territory_to_unit,
)
from diplomacy_cli.core.logic.validator.orchestrator import make_semantic_map


@pytest.fixture
def order_factory():
    def _factory(
        origin: str,
        order_type: OrderType,
        destination: str | None = None,
        support_origin: str | None = None,
        support_destination: str | None = None,
        convoy_origin: str | None = None,
        convoy_destination: str | None = None,
        unit_type: UnitType | None = None,
    ) -> Order:
        return Order(
            origin=origin,
            order_type=order_type,
            destination=destination,
            support_origin=support_origin,
            support_destination=support_destination,
            convoy_origin=convoy_origin,
            convoy_destination=convoy_destination,
            unit_type=unit_type,
        )

    return _factory


@pytest.fixture
def semantic_result_factory(order_factory):
    def _factory(
        player_id: str,
        origin: str,
        order_type: OrderType,
        valid: bool = True,
        errors: list[str] | None = None,
        raw: str = "",
        normalized: str = "",
        **order_kwargs,
    ) -> SemanticResult:
        order = order_factory(
            origin=origin,
            order_type=order_type,
            **order_kwargs,
        )
        return SemanticResult(
            player_id=player_id,
            raw=raw,
            normalized=normalized,
            order=order,
            valid=valid,
            errors=errors or [],
        )

    return _factory


@pytest.fixture
def semantic_map_factory(semantic_result_factory):
    def _factory(
        loaded_state,
        sem_kwargs_list,
    ) -> tuple[dict[str, SemanticResult], dict[str, list[SemanticResult]]]:
        semantic_results = [
            semantic_result_factory(**kwargs) for kwargs in sem_kwargs_list
        ]
        return make_semantic_map(loaded_state, semantic_results)

    return _factory


@pytest.fixture
def loaded_state_factory():
    def _factory(
        unit_specs: list[tuple[str, str, UnitType, str]],
        territory_state: dict[str, dict] | None = None,
        game_meta: dict | None = None,
        players: dict[str, dict] | None = None,
        raw_orders: dict[str, list[str]] = {},
        pending_move=None,
    ) -> LoadedState:
        players = players or {
            owner_id: {"status": "active", "nation_id": owner_id}
            for _, owner_id, _, _ in unit_specs
        }

        units = {
            uid: {
                "owner_id": owner_id,
                "unit_type": unit_type,
                "territory_id": territory,
            }
            for uid, owner_id, unit_type, territory in unit_specs
        }

        game = GameState(
            game_meta=game_meta
            or {
                "game_id": "test_game",
                "variant": "classic",
                "turn_code": "1901-S-M",
            },
            players=players,
            territory_state=territory_state or {},
            units=units,
            raw_orders=raw_orders,
        )

        return LoadedState(
            game=game,
            territory_to_unit=build_territory_to_unit(units),
            counters=build_counters(units),
            pending_move=pending_move,
        )

    return _factory


@pytest.fixture
def resolution_soa_factory():
    def _factory(
        unit_id: list[str],
        owner_id: list[str],
        unit_type: list[UnitType],
        orig_territory: list[str],
        order_type: list[OrderType],
        move_destination: list[str | None],
        support_origin: list[str | None],
        support_destination: list[str | None],
        convoy_origin: list[str | None],
        convoy_destination: list[str | None],
        new_territory: list[str] | None = None,
        strength: list[int] | None = None,
        dislodged: list[bool] | None = None,
        support_cut: list[bool] | None = None,
        convoy_path_flat: list[str] | None = None,
        convoy_path_start: list[int] | None = None,
        convoy_path_len: list[int] | None = None,
        outcome: list[OutcomeType | None] | None = None,
    ) -> ResolutionSoA:
        n = len(unit_id)

        def default(lst, default_val):
            return lst if lst is not None else [default_val] * n

        return ResolutionSoA(
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
            new_territory=cast(
                list[str], new_territory or orig_territory.copy()
            ),
            strength=default(strength, 1),
            dislodged=default(dislodged, False),
            support_cut=default(support_cut, False),
            convoy_path_flat=convoy_path_flat
            if convoy_path_flat is not None
            else [],
            convoy_path_len=convoy_path_len
            if convoy_path_len is not None
            else [0] * n,
            convoy_path_start=convoy_path_start
            if convoy_path_start is not None
            else [-1] * n,
            outcome=cast(list[OutcomeType | None], default(outcome, None)),
        )

    return _factory


@pytest.fixture
def rules_factory():
    def _factory(
        parent_to_coast: dict[str, set[str]],
        adjacency_map: dict[str, list[tuple[str, str]]],
        territory_ids: list[str] | None = None,
    ):
        ns = SimpleNamespace(
            parent_to_coast=parent_to_coast,
            adjacency_map=adjacency_map,
            territory_ids=territory_ids
            or list(parent_to_coast.keys()) + list(adjacency_map.keys()),
        )
        return cast(Rules, ns)

    return _factory


@pytest.fixture
def classic_rules():
    return load_rules("classic")
