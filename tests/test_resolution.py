from diplomacy_cli.core.logic.schema import OrderType, UnitType
from diplomacy_cli.core.logic.validator.resolution import (
    make_resolution_maps,
    make_semantic_map,
    move_phase_soa,
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
    assert soa.has_valid_convoy == [False, False]
    assert soa.is_resolved == [False, False]
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
