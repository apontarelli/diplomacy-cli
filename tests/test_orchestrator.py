from diplomacy_cli.core.logic.schema import OrderType, UnitType
from diplomacy_cli.core.logic.validator.orchestrator import make_semantic_map


def test_make_semantic_map_defaults_to_hold(loaded_state_factory):
    ls = loaded_state_factory(
        [
            ("U1", 1, UnitType.ARMY, "A"),
            ("U2", 2, UnitType.FLEET, "B"),
        ]
    )
    sem_by_unit, errs = make_semantic_map(ls, [])
    assert errs == {}
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
        raw="A-Y", origin="A", order_type=OrderType.MOVE, destination="Y"
    )
    sem_by_unit, errs = make_semantic_map(ls, [sem1, sem2])

    assert list(errs.keys()) == ["U1"]
    assert len(errs["U1"]) == 1
    assert errs["U1"][0].raw == "A-Y"  # check the SemanticResult content
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
    assert errs == {}
    assert sem_by_unit["U1"].order.order_type == OrderType.MOVE
    assert sem_by_unit["U1"].order.destination == "X"
    assert sem_by_unit["U2"].order.order_type == OrderType.SUPPORT_HOLD
    assert sem_by_unit["U2"].order.support_origin == "A"
