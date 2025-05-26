from types import SimpleNamespace
from typing import cast

import pytest

from diplomacy_cli.core.logic.schema import (
    Counters,
    GameState,
    LoadedState,
    Order,
    OrderType,
    OutcomeType,
    ResolutionSoA,
    Rules,
    SemanticResult,
    TerritoryToUnit,
    UnitType,
)


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
    ) -> Order:
        return Order(
            origin=origin,
            order_type=order_type,
            destination=destination,
            support_origin=support_origin,
            support_destination=support_destination,
            convoy_origin=convoy_origin,
            convoy_destination=convoy_destination,
        )

    return _factory


@pytest.fixture
def semantic_result_factory(order_factory):
    def _factory(
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
            raw=raw,
            normalized=normalized,
            order=order,
            valid=valid,
            errors=errors or [],
        )

    return _factory


@pytest.fixture
def loaded_state_factory():
    def _factory(
        unit_specs: list[tuple[str, int, UnitType, str]],
        counters: Counters | None = None,
        dislodged: set[str] | None = None,
    ) -> LoadedState:
        gs_ns = SimpleNamespace(
            game_meta=SimpleNamespace(),
            players=[],
            territory_state={},
            units={},
            raw_orders=[],
        )

        for uid, owner_id, utype, territory in unit_specs:
            gs_ns.units[uid] = {
                "owner_id": owner_id,
                "unit_type": utype,
                "territory_id": territory,
            }

        game = cast(GameState, gs_ns)

        territory_to_unit: TerritoryToUnit = {
            territory: uid for uid, _, _, territory in unit_specs
        }

        if counters is None:
            counters = {}
            for uid, _, _, _ in unit_specs:
                parts = uid.split("_")
                if len(parts) == 3:
                    key = f"{parts[0]}_{parts[1]}"
                    num = int(parts[2])
                    counters[key] = max(counters.get(key, 0), num)

        dislodged = dislodged or set()

        return LoadedState(
            game=game,
            territory_to_unit=territory_to_unit,
            counters=counters,
            dislodged=dislodged,
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
    ):
        ns = SimpleNamespace(
            parent_to_coast=parent_to_coast,
            adjacency_map=adjacency_map,
        )
        return cast(Rules, ns)

    return _factory
