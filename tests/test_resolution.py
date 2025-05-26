from types import SimpleNamespace
from typing import cast

import pytest

from diplomacy_cli.core.logic.schema import (
    OrderType,
    OutcomeType,
    ResolutionSoA,
    UnitType,
)
from diplomacy_cli.core.logic.validator.resolution import (
    ResolutionMaps,
    calculate_strength,
    cut_supports,
    detect_dislodged,
    find_convoy_path,
    flag_support_convoy_mismatches,
    get_convoy_path,
    make_resolution_maps,
    make_semantic_map,
    move_phase_soa,
    process_convoys,
    process_moves,
    resolve_conflict,
    assign_move_outcomes,
)


def test_make_resolution_maps_single_move(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        order_type=[OrderType.MOVE],
        move_destination=["B"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
    )
    maps = make_resolution_maps(soa)

    assert maps.move_by_origin == {"A": 0}
    assert maps.moves_by_dest == {"B": [0]}

    assert not maps.support_by_origin
    assert not maps.support_moves_by_supported_origin
    assert not maps.support_moves_by_supported_dest
    assert not maps.support_holds_by_supported_origin
    assert not maps.convoy_by_origin
    assert not maps.convoys_by_army_origin
    assert not maps.convoys_by_army_dest
    assert not maps.hold_by_origin


def test_make_resolution_maps_mixed_orders_ignore_none(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2", "u3"],
        owner_id=["p1", "p2", "p3"],
        unit_type=[UnitType.ARMY, UnitType.FLEET, UnitType.ARMY],
        orig_territory=["A", "B", "C"],
        order_type=[OrderType.MOVE, OrderType.MOVE, OrderType.HOLD],
        move_destination=["X", "Y", None],
        support_origin=[None, None, None],
        support_destination=[None, None, None],
        convoy_origin=[None, None, None],
        convoy_destination=[None, None, None],
    )
    maps = make_resolution_maps(soa)

    assert maps.move_by_origin == {"A": 0, "B": 1}
    assert set(maps.moves_by_dest.keys()) == {"X", "Y"}
    assert maps.moves_by_dest["X"] == [0]
    assert maps.moves_by_dest["Y"] == [1]

    assert maps.hold_by_origin == {"C": 2}

    assert not maps.support_by_origin
    assert not maps.support_moves_by_supported_origin
    assert not maps.support_moves_by_supported_dest
    assert not maps.support_holds_by_supported_origin
    assert not maps.convoy_by_origin
    assert not maps.convoys_by_army_origin
    assert not maps.convoys_by_army_dest


def test_make_resolution_maps_no_orders(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=[],
        owner_id=[],
        unit_type=[],
        orig_territory=[],
        order_type=[],
        move_destination=[],
        support_origin=[],
        support_destination=[],
        convoy_origin=[],
        convoy_destination=[],
    )
    maps = make_resolution_maps(soa)

    assert not maps.move_by_origin
    assert not maps.moves_by_dest
    assert not maps.support_by_origin
    assert not maps.support_moves_by_supported_origin
    assert not maps.support_moves_by_supported_dest
    assert not maps.support_holds_by_supported_origin
    assert not maps.convoy_by_origin
    assert not maps.convoys_by_army_origin
    assert not maps.convoys_by_army_dest
    assert not maps.hold_by_origin


def test_make_semantic_map_defaults_to_hold(loaded_state_factory):
    ls = loaded_state_factory(
        [
            ("U1", 1, UnitType.ARMY, "A"),
            ("U2", 2, UnitType.FLEET, "B"),
        ]
    )
    sem_by_unit, errs = make_semantic_map(ls, [])
    assert errs == []
    assert set(sem_by_unit.keys()) == {"U1", "U2"}
    for sem in sem_by_unit.values():
        assert sem.order.order_type == OrderType.HOLD
        assert sem.valid


def test_make_semantic_map_duplicate_orders(
    loaded_state_factory, semantic_result_factory
):
    ls = loaded_state_factory(
        [
            ("U1", 1, UnitType.ARMY, "A"),
            ("U2", 2, UnitType.FLEET, "B"),
        ]
    )
    sem1 = semantic_result_factory(
        origin="A", order_type=OrderType.MOVE, destination="X"
    )
    sem2 = semantic_result_factory(
        origin="A", order_type=OrderType.MOVE, destination="Y"
    )
    sem_by_unit, errs = make_semantic_map(ls, [sem1, sem2])
    assert errs == ["Duplicate order for unit U1; ignoring."]
    assert sem_by_unit["U1"].order.destination == "X"
    assert sem_by_unit["U2"].order.order_type == OrderType.HOLD


def test_make_semantic_map_preexisting(
    loaded_state_factory, semantic_result_factory
):
    ls = loaded_state_factory(
        [
            ("U1", 1, UnitType.ARMY, "A"),
            ("U2", 2, UnitType.FLEET, "B"),
        ]
    )
    move_sem = semantic_result_factory(
        origin="A", order_type=OrderType.MOVE, destination="X"
    )
    support_sem = semantic_result_factory(
        origin="B",
        order_type=OrderType.SUPPORT_HOLD,
        support_origin="A",
        support_destination="C",
    )
    sem_by_unit, errs = make_semantic_map(ls, [move_sem, support_sem])
    assert errs == []
    assert sem_by_unit["U1"].order.order_type == OrderType.MOVE
    assert sem_by_unit["U1"].order.destination == "X"
    assert sem_by_unit["U2"].order.order_type == OrderType.SUPPORT_HOLD
    assert sem_by_unit["U2"].order.support_origin == "A"


def test_move_phase_soa_basic(loaded_state_factory, semantic_result_factory):
    ls = loaded_state_factory(
        [
            ("U1", 1, UnitType.ARMY, "A"),
            ("U2", 2, UnitType.FLEET, "B"),
        ]
    )
    sems = [
        semantic_result_factory(
            origin="A", order_type=OrderType.MOVE, destination="C"
        ),
        semantic_result_factory(origin="B", order_type=OrderType.HOLD),
    ]
    soa, errs = move_phase_soa(ls, sems)
    assert errs == []
    assert soa.order_type == [OrderType.MOVE, OrderType.HOLD]
    assert soa.move_destination == ["C", None]
    assert soa.new_territory == ["A", "B"]
    assert soa.strength == [1, 1]
    assert soa.dislodged == [False, False]
    assert soa.support_cut == [False, False]
    assert soa.convoy_path_start == [-1, -1]
    assert soa.convoy_path_len == [0, 0]
    assert soa.convoy_path_flat == []
    assert soa.outcome == [None, None]


def test_move_phase_soa_missing_data(
    loaded_state_factory, semantic_result_factory
):
    ls = loaded_state_factory(
        [
            ("U1", 1, UnitType.ARMY, "A"),
            ("U2", 2, UnitType.FLEET, "B"),
        ]
    )
    sems = [
        semantic_result_factory(
            origin="A", order_type=OrderType.MOVE, destination="C"
        )
    ]
    soa, errs = move_phase_soa(ls, sems)
    assert errs == []
    assert soa.order_type == [OrderType.MOVE, OrderType.HOLD]
    assert soa.move_destination == ["C", None]
    assert soa.new_territory == ["A", "B"]


def test_valid_convoy_support(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.FLEET],
        orig_territory=["A", "B"],
        order_type=[OrderType.MOVE, OrderType.CONVOY],
        move_destination=["C", None],
        support_origin=[None, None],
        support_destination=[None, None],
        convoy_origin=[None, "A"],
        convoy_destination=[None, "C"],
    )
    maps = make_resolution_maps(soa)
    outcomes = flag_support_convoy_mismatches(soa, maps)
    assert outcomes == [None, None]


def test_valid_support_move(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["A", "B"],
        order_type=[OrderType.MOVE, OrderType.SUPPORT_MOVE],
        move_destination=["X", None],
        support_origin=[None, "A"],
        support_destination=[None, "X"],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
    )
    maps = make_resolution_maps(soa)
    outcomes = flag_support_convoy_mismatches(soa, maps)
    assert outcomes == [None, None]


def test_valid_support_hold(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["A", "B"],
        order_type=[OrderType.HOLD, OrderType.SUPPORT_HOLD],
        move_destination=[None, None],
        support_origin=[None, "A"],
        support_destination=[None, "Z"],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
    )
    maps = make_resolution_maps(soa)
    outcomes = flag_support_convoy_mismatches(soa, maps)
    assert outcomes == [None, None]


def test_invalid_convoy_wrong_dest(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.FLEET],
        orig_territory=["A", "B"],
        order_type=[OrderType.MOVE, OrderType.CONVOY],
        move_destination=["C", None],
        support_origin=[None, None],
        support_destination=[None, None],
        convoy_origin=[None, "A"],
        convoy_destination=[None, "D"],
    )
    maps = make_resolution_maps(soa)
    outcomes = flag_support_convoy_mismatches(soa, maps)
    assert outcomes == [None, OutcomeType.INVALID_CONVOY]


def test_invalid_support_move(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["A", "B"],
        order_type=[OrderType.MOVE, OrderType.SUPPORT_MOVE],
        move_destination=["X", None],
        support_origin=[None, "A"],
        support_destination=[None, "Y"],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
    )
    maps = make_resolution_maps(soa)
    outcomes = flag_support_convoy_mismatches(soa, maps)
    assert outcomes == [None, OutcomeType.INVALID_SUPPORT]


def test_invalid_support_hold_wrong_type(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["A", "B"],
        order_type=[
            OrderType.MOVE,
            OrderType.SUPPORT_HOLD,
        ],
        move_destination=["C", None],
        support_origin=[None, "A"],
        support_destination=[None, "C"],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
    )
    maps = make_resolution_maps(soa)
    outcomes = flag_support_convoy_mismatches(soa, maps)
    assert outcomes == [None, OutcomeType.INVALID_SUPPORT]


def test_find_convoy_path_basic(rules_factory):
    rules = rules_factory(
        parent_to_coast={"A": {"A"}, "B": {"B"}, "C": {"C"}},
        adjacency_map={
            "A": [("B", "sea")],
            "B": [("A", "sea"), ("C", "sea")],
            "C": [("B", "sea")],
        },
    )
    result = find_convoy_path("A", "C", ["B"], rules)
    assert result == ["A", "B", "C"]


def test_find_convoy_path_disconnected(rules_factory):
    rules = rules_factory(
        parent_to_coast={"A": {"A"}, "C": {"C"}},
        adjacency_map={
            "A": [("X", "sea")],
            "X": [("A", "sea")],
            "C": [("Y", "sea")],
            "Y": [("C", "sea")],
        },
    )
    result = find_convoy_path("A", "C", ["X", "Y"], rules)
    assert result is None


def test_get_convoy_path_success():
    ns = SimpleNamespace(
        convoy_path_start=[0],
        convoy_path_len=[3],
        convoy_path_flat=["A", "B", "C"],
    )
    soa = cast(ResolutionSoA, ns)
    result = get_convoy_path(soa, 0)
    assert result == ["A", "B", "C"]


def test_get_convoy_path_none():
    ns = SimpleNamespace(
        convoy_path_start=[-1],
        convoy_path_len=[0],
        convoy_path_flat=[],
    )
    soa = cast(ResolutionSoA, ns)
    result = get_convoy_path(soa, 0)
    assert result is None


def test_process_convoys_valid_path(resolution_soa_factory, rules_factory):
    rules = rules_factory(
        parent_to_coast={"A": {"A"}, "C": {"C"}},
        adjacency_map={
            "A": [("B", "sea")],
            "B": [("A", "sea"), ("C", "sea")],
            "C": [("B", "sea")],
        },
    )

    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=["army", "fleet"],
        orig_territory=["A", "B"],
        order_type=["move", "convoy"],
        move_destination=["C", None],
        support_origin=[None, None],
        support_destination=[None, None],
        convoy_origin=[None, "A"],
        convoy_destination=[None, "C"],
        dislodged=[False, False],
        outcome=[None, None],
        convoy_path_flat=[],
    )

    origin_to_move = {"A": 0}
    origin_to_convoy = {"B": 1}

    path_start, path_len, path_flat = process_convoys(
        soa, rules, origin_to_move, origin_to_convoy
    )

    assert path_start[0] == 0
    assert path_len[0] == 3
    assert path_flat == ["A", "B", "C"]


def test_process_convoys_invalid_convoy(resolution_soa_factory, rules_factory):
    rules = rules_factory(
        parent_to_coast={},
        adjacency_map={},
    )

    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=["army", "fleet"],
        orig_territory=["A", "B"],
        order_type=["move", "convoy"],
        move_destination=["C", None],
        support_origin=[None, None],
        support_destination=[None, None],
        convoy_origin=[None, "A"],
        convoy_destination=[None, "C"],
        dislodged=[False, True],
        outcome=[None, None],
        convoy_path_flat=[],
    )

    origin_to_move = {"A": 0}
    origin_to_convoy = {"B": 1}

    path_start, path_len, path_flat = process_convoys(
        soa, rules, origin_to_move, origin_to_convoy
    )

    assert path_start == [-1] * len(soa.order_type)
    assert path_len == [0] * len(soa.order_type)
    assert path_flat == []


def test_process_moves_army_adjacent(resolution_soa_factory, rules_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        order_type=[OrderType.MOVE],
        move_destination=["B"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
    )
    rules = rules_factory(
        parent_to_coast={},
        adjacency_map={"A": [("B", "land")]},
    )
    new_territory, outcome = process_moves(soa, {"A": 0}, rules)
    assert new_territory == ["B"]
    assert outcome == [None]


def test_process_moves_army_with_convoy(resolution_soa_factory, rules_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        order_type=[OrderType.MOVE],
        move_destination=["C"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
        convoy_path_len=[2],
    )
    rules = rules_factory(
        parent_to_coast={},
        adjacency_map={"A": [("B", "land")]},
    )
    new_territory, outcome = process_moves(soa, {"A": 0}, rules)
    assert new_territory == ["C"]
    assert outcome == [None]


def test_process_moves_army_invalid_no_convoy(
    resolution_soa_factory, rules_factory
):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        order_type=[OrderType.MOVE],
        move_destination=["C"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
        convoy_path_len=[0],
    )
    rules = rules_factory(
        parent_to_coast={},
        adjacency_map={"A": [("B", "land")]},
    )
    new_territory, outcome = process_moves(soa, {"A": 0}, rules)
    assert new_territory == ["A"]
    assert outcome == [OutcomeType.MOVE_NO_CONVOY]


def test_process_moves_fleet_adjacent(resolution_soa_factory, rules_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.FLEET],
        orig_territory=["X"],
        order_type=[OrderType.MOVE],
        move_destination=["Y"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
    )
    rules = rules_factory(
        parent_to_coast={},
        adjacency_map={"X": [("Y", "sea")]},
    )
    new_territory, outcome = process_moves(soa, {"X": 0}, rules)
    assert new_territory == ["Y"]
    assert outcome == [None]


def test_process_moves_fleet_invalid_not_adjacent(
    resolution_soa_factory, rules_factory
):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.FLEET],
        orig_territory=["X"],
        order_type=[OrderType.MOVE],
        move_destination=["Z"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
    )
    rules = rules_factory(
        parent_to_coast={},
        adjacency_map={"X": [("Y", "sea")]},
    )
    new_territory, outcome = process_moves(soa, {"X": 0}, rules)
    assert new_territory == ["X"]
    assert outcome == [None]


def test_cut_support_success_support_hold_attacked_by_third_party(
    resolution_soa_factory,
):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2", "u3"],
        owner_id=["p1", "p2", "p3"],
        unit_type=[UnitType.ARMY] * 3,
        orig_territory=["A", "B", "C"],
        order_type=[OrderType.SUPPORT_HOLD, OrderType.MOVE, OrderType.HOLD],
        move_destination=[None, "A", None],
        support_origin=[None, None, None],
        support_destination=["C", None, None],
        convoy_origin=[None] * 3,
        convoy_destination=[None] * 3,
        new_territory=["A", "A", "C"],
        outcome=[None, OutcomeType.MOVE_SUCCESS, None],
    )
    move_by_origin = {"B": 1}
    result = cut_supports(soa, move_by_origin)
    assert result == [True, False, False]


def test_support_hold_not_cut_if_supporting_unit_attacks_supporter(
    resolution_soa_factory,
):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["B", "A"],
        order_type=[OrderType.SUPPORT_HOLD, OrderType.MOVE],
        move_destination=[None, "B"],
        support_origin=[None, None],
        support_destination=["A", None],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        new_territory=["B", "B"],
        outcome=[None, OutcomeType.MOVE_SUCCESS],
    )
    move_by_origin = {"A": 1}
    result = cut_supports(soa, move_by_origin)
    assert result == [False, False]


def test_cut_support_fails_wrong_destination(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["B", "A"],
        order_type=[OrderType.SUPPORT_HOLD, OrderType.MOVE],
        move_destination=[None, "C"],
        support_origin=[None, None],
        support_destination=["A", None],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        new_territory=["B", "C"],
        outcome=[None, OutcomeType.MOVE_SUCCESS],
    )
    move_by_origin = {"A": 1}
    result = cut_supports(soa, move_by_origin)
    assert result == [False, False]


def test_cut_support_fails_move_no_convoy(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["B", "A"],
        order_type=[OrderType.SUPPORT_HOLD, OrderType.MOVE],
        move_destination=[None, "B"],
        support_origin=[None, None],
        support_destination=["A", None],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        new_territory=["B", "B"],
        outcome=[None, OutcomeType.MOVE_NO_CONVOY],
    )
    move_by_origin = {"A": 1}
    result = cut_supports(soa, move_by_origin)
    assert result == [False, False]


def test_cut_support_fails_from_supported_unit(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["B", "A"],
        order_type=[OrderType.SUPPORT_HOLD, OrderType.MOVE],
        move_destination=[None, "B"],
        support_origin=[None, None],
        support_destination=["A", None],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        new_territory=["B", "B"],
        outcome=[None, OutcomeType.MOVE_SUCCESS],
    )
    move_by_origin = {"A": 1}
    result = cut_supports(soa, move_by_origin)
    assert result == [False, False]


def test_non_support_order_ignored(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY, UnitType.ARMY],
        orig_territory=["B", "A"],
        order_type=[OrderType.MOVE, OrderType.MOVE],
        move_destination=["C", "B"],
        support_origin=[None, None],
        support_destination=[None, None],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        new_territory=["C", "B"],
        outcome=[OutcomeType.MOVE_SUCCESS, OutcomeType.MOVE_SUCCESS],
    )
    move_by_origin = {"A": 1}
    result = cut_supports(soa, move_by_origin)
    assert result == [False, False]


def test_strength_no_support(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY] * 2,
        orig_territory=["A", "B"],
        order_type=[OrderType.HOLD, OrderType.HOLD],
        move_destination=[None, None],
        support_origin=[None, None],
        support_destination=[None, None],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        support_cut=[False, False],
    )

    maps = cast(
        ResolutionMaps,
        SimpleNamespace(
            move_by_origin={},
            hold_by_origin={"A": 0, "B": 1},
            support_moves_by_supported_origin={},
            support_holds_by_supported_origin={},
        ),
    )

    result = calculate_strength(soa, maps)
    assert result == [1, 1]


def test_strength_with_uncut_support_hold(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY] * 2,
        orig_territory=["A", "B"],
        order_type=[OrderType.HOLD, OrderType.SUPPORT_HOLD],
        move_destination=[None, None],
        support_origin=[None, "A"],
        support_destination=[None, None],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        support_cut=[False, False],
    )

    maps = cast(
        ResolutionMaps,
        SimpleNamespace(
            move_by_origin={},
            hold_by_origin={"A": 0, "B": 1},
            support_moves_by_supported_origin={},
            support_holds_by_supported_origin={"A": [1]},
        ),
    )

    result = calculate_strength(soa, maps)
    assert result == [2, 1]


def test_strength_with_uncut_support_move(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY] * 2,
        orig_territory=["A", "B"],
        order_type=[OrderType.MOVE, OrderType.SUPPORT_MOVE],
        move_destination=["C", None],
        support_origin=[None, "A"],
        support_destination=[None, "C"],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        support_cut=[False, False],
    )

    maps = cast(
        ResolutionMaps,
        SimpleNamespace(
            move_by_origin={"A": 0},
            hold_by_origin={},
            support_moves_by_supported_origin={"A": [1]},
            support_holds_by_supported_origin={},
        ),
    )

    result = calculate_strength(soa, maps)
    assert result == [2, 1]


def test_strength_multiple_uncut_supports(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2", "u3"],
        owner_id=["p1", "p2", "p3"],
        unit_type=[UnitType.ARMY] * 3,
        orig_territory=["A", "B", "C"],
        order_type=[
            OrderType.HOLD,
            OrderType.SUPPORT_HOLD,
            OrderType.SUPPORT_HOLD,
        ],
        move_destination=[None, None, None],
        support_origin=[None, "A", "A"],
        support_destination=[None, None, None],
        convoy_origin=[None, None, None],
        convoy_destination=[None, None, None],
        support_cut=[False, False, False],
    )

    maps = cast(
        ResolutionMaps,
        SimpleNamespace(
            move_by_origin={},
            hold_by_origin={"A": 0, "B": 1, "C": 2},
            support_moves_by_supported_origin={},
            support_holds_by_supported_origin={"A": [1, 2]},
        ),
    )

    result = calculate_strength(soa, maps)
    assert result == [3, 1, 1]


def test_strength_support_cut_ignored(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY] * 2,
        orig_territory=["A", "B"],
        order_type=[OrderType.HOLD, OrderType.SUPPORT_HOLD],
        move_destination=[None, None],
        support_origin=[None, None],
        support_destination=[None, "A"],
        convoy_origin=[None, None],
        convoy_destination=[None, None],
        support_cut=[False, True],
    )

    maps = cast(
        ResolutionMaps,
        SimpleNamespace(
            move_by_origin={},
            hold_by_origin={"A": 0, "B": 1},
            support_moves_by_supported_origin={},
            support_holds_by_supported_origin={"A": [1]},
        ),
    )

    result = calculate_strength(soa, maps)
    assert result == [1, 1]


def test_strength_mixed_support_move_and_hold(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2", "u3", "u4"],
        owner_id=["p1"] * 4,
        unit_type=[UnitType.ARMY] * 4,
        orig_territory=["A", "B", "C", "D"],
        order_type=[
            OrderType.MOVE,
            OrderType.SUPPORT_MOVE,
            OrderType.HOLD,
            OrderType.SUPPORT_HOLD,
        ],
        move_destination=["X", None, None, None],
        support_origin=[None, "A", None, "C"],
        support_destination=[None, "X", None, None],
        convoy_origin=[None] * 4,
        convoy_destination=[None] * 4,
        support_cut=[False] * 4,
    )

    maps = cast(
        ResolutionMaps,
        SimpleNamespace(
            move_by_origin={"A": 0},
            hold_by_origin={"B": 1, "C": 2, "D": 3},
            support_moves_by_supported_origin={"A": [1]},
            support_holds_by_supported_origin={"C": [3]},
        ),
    )

    result = calculate_strength(soa, maps)
    assert result == [2, 1, 2, 1]


def test_resolve_no_conflict(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY] * 2,
        orig_territory=["A", "B"],
        new_territory=["A", "C"],
        order_type=[OrderType.HOLD, OrderType.MOVE],
        move_destination=[None, "C"],
        support_origin=[None] * 2,
        support_destination=[None] * 2,
        convoy_origin=[None] * 2,
        convoy_destination=[None] * 2,
        strength=[1, 1],
    )
    result = resolve_conflict(soa)
    assert result == ["A", "C"]


def test_resolve_clear_winner(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY] * 2,
        orig_territory=["A", "B"],
        new_territory=["C", "C"],
        order_type=[OrderType.MOVE] * 2,
        move_destination=["C", "C"],
        support_origin=[None] * 2,
        support_destination=[None] * 2,
        convoy_origin=[None] * 2,
        convoy_destination=[None] * 2,
        strength=[2, 1],
    )
    result = resolve_conflict(soa)
    assert result == ["C", "B"]


def test_resolve_tie_bounce(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2"],
        owner_id=["p1", "p2"],
        unit_type=[UnitType.ARMY] * 2,
        orig_territory=["A", "B"],
        new_territory=["C", "C"],
        order_type=[OrderType.MOVE] * 2,
        move_destination=["C", "C"],
        support_origin=[None] * 2,
        support_destination=[None] * 2,
        convoy_origin=[None] * 2,
        convoy_destination=[None] * 2,
        strength=[2, 2],
    )
    result = resolve_conflict(soa)
    assert result == ["A", "B"]


def test_resolve_cascading_conflict_corrected(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2", "u3"],
        owner_id=["p1"] * 3,
        unit_type=[UnitType.ARMY] * 3,
        orig_territory=["A", "B", "C"],
        order_type=[OrderType.MOVE, OrderType.MOVE, OrderType.HOLD],
        move_destination=["B", "C", None],
        new_territory=["B", "C", "C"],
        support_origin=[None] * 3,
        support_destination=[None] * 3,
        convoy_origin=[None] * 3,
        convoy_destination=[None] * 3,
        strength=[2, 2, 1],
    )
    result = resolve_conflict(soa)
    assert result == ["B", "C", "C"]


def test_resolve_cascading_tie(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2", "u3"],
        owner_id=["p1"] * 3,
        unit_type=[UnitType.ARMY] * 3,
        orig_territory=["A", "C", "D"],
        order_type=[OrderType.MOVE] * 3,
        move_destination=["B", "B", "C"],
        new_territory=["B", "B", "C"],
        support_origin=[None] * 3,
        support_destination=[None] * 3,
        convoy_origin=[None] * 3,
        convoy_destination=[None] * 3,
        strength=[2, 2, 2],
    )

    result = resolve_conflict(soa)
    assert result == ["A", "C", "D"]


@pytest.mark.parametrize(
    "description, soa_kwargs, expected",
    [
        (
            "Hold dislodged by enemy move",
            {
                "unit_id": ["u1", "u2"],
                "owner_id": ["p1", "p2"],
                "unit_type": [UnitType.ARMY, UnitType.ARMY],
                "orig_territory": ["A", "B"],
                "order_type": [OrderType.HOLD, OrderType.MOVE],
                "move_destination": [None, "A"],
                "support_origin": [None, None],
                "support_destination": [None, None],
                "convoy_origin": [None, None],
                "convoy_destination": [None, None],
                "new_territory": ["A", "A"],
            },
            [True, False],
        ),
        (
            "Bounce prevents dislodgement",
            {
                "unit_id": ["u1", "u2", "u3"],
                "owner_id": ["p1", "p2", "p3"],
                "unit_type": [UnitType.ARMY] * 3,
                "orig_territory": ["A", "B", "C"],
                "order_type": [OrderType.HOLD, OrderType.MOVE, OrderType.MOVE],
                "move_destination": [None, "A", "A"],
                "support_origin": [None, None, None],
                "support_destination": [None, None, None],
                "convoy_origin": [None, None, None],
                "convoy_destination": [None, None, None],
                "new_territory": ["A", "B", "C"],
            },
            [False, False, False],
        ),
        (
            "Bounced unit dislodged",
            {
                "unit_id": ["u1", "u2"],
                "owner_id": ["p1", "p2"],
                "unit_type": [UnitType.ARMY, UnitType.ARMY],
                "orig_territory": ["A", "B"],
                "order_type": [OrderType.MOVE, OrderType.MOVE],
                "move_destination": ["B", "A"],
                "support_origin": [None, None],
                "support_destination": [None, None],
                "convoy_origin": [None, None],
                "convoy_destination": [None, None],
                "new_territory": ["A", "A"],
            },
            [True, False],
        ),
        (
            "Self dislodge prevented",
            {
                "unit_id": ["u1", "u2"],
                "owner_id": ["p1", "p1"],
                "unit_type": [UnitType.ARMY, UnitType.ARMY],
                "orig_territory": ["A", "B"],
                "order_type": [OrderType.HOLD, OrderType.MOVE],
                "move_destination": [None, "A"],
                "support_origin": [None, None],
                "support_destination": [None, None],
                "convoy_origin": [None, None],
                "convoy_destination": [None, None],
                "new_territory": ["A", "A"],
            },
            [False, False],
        ),
    ],
)
def test_detect_dislodged(
    resolution_soa_factory, description, soa_kwargs, expected
):
    soa = resolution_soa_factory(**soa_kwargs)
    result = detect_dislodged(soa)
    assert result == expected, f"Failed: {description}"


def test_assign_move_success(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        new_territory=["B"],
        order_type=[OrderType.MOVE],
        move_destination=["B"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
    )
    outcome = assign_move_outcomes(soa)
    assert outcome == [OutcomeType.MOVE_SUCCESS]


def test_assign_move_bounced(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        new_territory=["A"],
        order_type=[OrderType.MOVE],
        move_destination=["B"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
    )
    outcome = assign_move_outcomes(soa)
    assert outcome == [OutcomeType.MOVE_BOUNCED]


def test_assign_move_inconsistent_raises(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        new_territory=["X"],  # invalid: not origin or destination
        order_type=[OrderType.MOVE],
        move_destination=["B"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
    )
    with pytest.raises(ValueError, match="Inconsistent new_territory"):
        assign_move_outcomes(soa)


def test_assign_support_cut(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        new_territory=["A"],
        order_type=[OrderType.SUPPORT_HOLD],
        move_destination=[None],
        support_origin=[None],
        support_destination=["X"],
        convoy_origin=[None],
        convoy_destination=[None],
        support_cut=[True],
    )
    outcome = assign_move_outcomes(soa)
    assert outcome == [OutcomeType.SUPPORT_CUT]


def test_assign_support_success(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        new_territory=["A"],
        order_type=[OrderType.SUPPORT_MOVE],
        move_destination=[None],
        support_origin=["X"],
        support_destination=["Y"],
        convoy_origin=[None],
        convoy_destination=[None],
        support_cut=[False],
    )
    outcome = assign_move_outcomes(soa)
    assert outcome == [OutcomeType.SUPPORT_SUCCESS]


def test_assign_hold_success(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        new_territory=["A"],
        order_type=[OrderType.HOLD],
        move_destination=[None],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
    )
    outcome = assign_move_outcomes(soa)
    assert outcome == [OutcomeType.HOLD_SUCCESS]


def test_assign_convoy_success(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.FLEET],
        orig_territory=["A"],
        new_territory=["A"],
        order_type=[OrderType.CONVOY],
        move_destination=[None],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=["X"],
        convoy_destination=["Y"],
    )
    outcome = assign_move_outcomes(soa)
    assert outcome == [OutcomeType.CONVOY_SUCCESS]


def test_assign_dislodged_overrides_all(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        new_territory=["A"],
        order_type=[OrderType.HOLD],
        move_destination=[None],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
        dislodged=[True],
    )
    outcome = assign_move_outcomes(soa)
    assert outcome == [OutcomeType.DISLODGED]


def test_preserve_existing_outcome(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1"],
        owner_id=["p1"],
        unit_type=[UnitType.ARMY],
        orig_territory=["A"],
        new_territory=["B"],
        order_type=[OrderType.MOVE],
        move_destination=["B"],
        support_origin=[None],
        support_destination=[None],
        convoy_origin=[None],
        convoy_destination=[None],
        outcome=[OutcomeType.MOVE_NO_CONVOY],
    )
    outcome = assign_move_outcomes(soa)
    assert outcome == [OutcomeType.MOVE_NO_CONVOY]


def test_assign_outcomes_mixed_batch(resolution_soa_factory):
    soa = resolution_soa_factory(
        unit_id=["u1", "u2", "u3", "u4"],
        owner_id=["p1"] * 4,
        unit_type=[UnitType.ARMY] * 4,
        orig_territory=["A", "B", "C", "D"],
        new_territory=["B", "B", "C", "D"],
        order_type=[
            OrderType.MOVE,
            OrderType.MOVE,
            OrderType.SUPPORT_HOLD,
            OrderType.HOLD,
        ],
        move_destination=["B", "C", None, None],
        support_origin=[None, None, None, None],
        support_destination=[None, None, "C", None],
        convoy_origin=[None] * 4,
        convoy_destination=[None] * 4,
        dislodged=[False, False, False, False],
        support_cut=[False, False, True, False],
        outcome=[None] * 4,
        strength=[2, 1, 1, 1],
    )
    result = assign_move_outcomes(soa)
    assert result == [
        OutcomeType.MOVE_SUCCESS,
        OutcomeType.MOVE_BOUNCED,
        OutcomeType.SUPPORT_CUT,
        OutcomeType.HOLD_SUCCESS,
    ]
