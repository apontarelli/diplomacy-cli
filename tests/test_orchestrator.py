from diplomacy_cli.core.logic.rules_loader import load_rules
from diplomacy_cli.core.logic.schema import (
    OrderType,
    OutcomeType,
    Phase,
    Season,
    UnitType,
)
from diplomacy_cli.core.logic.validator.orchestrator import (
    make_semantic_map,
    process_move_phase,
)


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
        player_id="eng", origin="A", order_type=OrderType.MOVE, destination="X"
    )
    sem2 = semantic_result_factory(
        player_id="eng",
        raw="A-Y",
        origin="A",
        order_type=OrderType.MOVE,
        destination="Y",
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
        player_id="eng", origin="A", order_type=OrderType.MOVE, destination="X"
    )
    support_sem = semantic_result_factory(
        player_id="eng",
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


def test_process_move_phase_single_valid_order(loaded_state_factory):
    raw_orders = {"P1": ["lon-wal"]}

    loaded_state = loaded_state_factory(
        [("U1", "P1", UnitType.ARMY, "lon")],
        game_meta={"turn_code": "1901-S-M"},
    )
    rules = load_rules("classic")
    report = process_move_phase(raw_orders, rules, loaded_state)

    assert report.phase == Phase.MOVEMENT
    assert report.year == 0
    assert report.season == Season.SPRING
    assert len(report.valid_syntax) == 1
    assert len(report.valid_semantics) == 1
    assert len(report.syntax_errors) == 0
    assert len(report.semantic_errors) == 0
    assert len(report.resolution_results) == 1

    res = report.resolution_results[0]
    assert res.unit_id == "U1"
    assert res.origin_territory == "lon"
    assert res.destination == "wal"
    assert res.outcome == OutcomeType.MOVE_SUCCESS


def test_process_move_phase_invalid_syntax(loaded_state_factory):
    raw_orders = {"P1": ["this-is-not-valid"]}

    loaded_state = loaded_state_factory(
        [("U1", "P1", UnitType.ARMY, "lon")],
        game_meta={"turn_code": "1901-S-M"},
    )

    rules = load_rules("classic")
    report = process_move_phase(raw_orders, rules, loaded_state)

    assert len(report.valid_syntax) == 0
    assert len(report.valid_semantics) == 0
    assert len(report.syntax_errors) == 1
    assert len(report.semantic_errors) == 0
    assert len(report.resolution_results) == 1


def test_process_move_phase_invalid_semantic(loaded_state_factory):
    raw_orders = {"P1": ["lon-mun"]}

    loaded_state = loaded_state_factory(
        [("U1", "P1", UnitType.ARMY, "lon")],
        game_meta={"turn_code": "1901-S-M"},
    )

    rules = load_rules("classic")

    report = process_move_phase(raw_orders, rules, loaded_state)

    assert len(report.valid_syntax) == 1
    assert len(report.valid_semantics) == 0
    assert len(report.syntax_errors) == 0
    assert len(report.semantic_errors) == 1
    assert len(report.resolution_results) == 1
