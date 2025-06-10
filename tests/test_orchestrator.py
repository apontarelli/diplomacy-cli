from diplomacy_cli.core.logic.schema import (
    OrderType,
    OutcomeType,
    Phase,
    Season,
    UnitType,
)
from diplomacy_cli.core.logic.validator.orchestrator import (
    make_adjustment_semantic_map,
    make_semantic_map,
    process_phase,
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
    assert errs["U1"][0].raw == "A-Y"
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


def test_make_adjustment_semantic_map_single_valid(
    loaded_state_factory, semantic_result_factory
):
    loaded_state = loaded_state_factory([("U1", "P1", UnitType.ARMY, "mun")])

    sem_results = [
        semantic_result_factory(
            player_id="P1",
            origin="mun",
            order_type=OrderType.DISBAND,
        ),
        semantic_result_factory(
            player_id="P1",
            origin="ber",
            order_type=OrderType.BUILD,
            unit_type=UnitType.ARMY,
        ),
    ]

    (
        disband_by_unit,
        dup_disband,
        build_by_territory,
        dup_build,
    ) = make_adjustment_semantic_map(loaded_state, sem_results)

    assert disband_by_unit == {"U1": sem_results[0]}
    assert dup_disband == {}
    assert build_by_territory == {"ber": sem_results[1]}
    assert dup_build == {}


def test_make_adjustment_semantic_map_duplicate_disband(
    loaded_state_factory, semantic_result_factory
):
    loaded_state = loaded_state_factory([("U1", "P1", UnitType.ARMY, "mun")])

    sem1 = semantic_result_factory("P1", "mun", OrderType.DISBAND)
    sem2 = semantic_result_factory("P1", "mun", OrderType.DISBAND)

    (
        disband_by_unit,
        dup_disband,
        build_by_territory,
        dup_build,
    ) = make_adjustment_semantic_map(loaded_state, [sem1, sem2])

    assert disband_by_unit == {"U1": sem1}
    assert dup_disband == {"U1": [sem2]}
    assert build_by_territory == {}
    assert dup_build == {}


def test_make_adjustment_semantic_map_duplicate_build(
    loaded_state_factory, semantic_result_factory
):
    loaded_state = loaded_state_factory([])

    sem1 = semantic_result_factory(
        "P1", "ber", OrderType.BUILD, unit_type=UnitType.ARMY
    )
    sem2 = semantic_result_factory(
        "P1", "ber", OrderType.BUILD, unit_type=UnitType.FLEET
    )

    (
        disband_by_unit,
        dup_disband,
        build_by_territory,
        dup_build,
    ) = make_adjustment_semantic_map(loaded_state, [sem1, sem2])

    assert disband_by_unit == {}
    assert dup_disband == {}
    assert build_by_territory == {"ber": sem1}
    assert dup_build == {"ber": [sem2]}


def test_process_move_phase_single_valid_order(
    loaded_state_factory, classic_rules
):
    raw_orders = {"P1": ["lon-wal"]}
    loaded_state = loaded_state_factory(
        [("U1", "P1", UnitType.ARMY, "lon")],
        game_meta={"turn_code": "1901-S-M"},
    )
    report = process_phase(raw_orders, classic_rules, loaded_state)

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


def test_process_move_phase_invalid_syntax(loaded_state_factory, classic_rules):
    raw_orders = {"P1": ["this-is-not-valid"]}
    loaded_state = loaded_state_factory(
        [("U1", "P1", UnitType.ARMY, "lon")],
        game_meta={"turn_code": "1901-S-M"},
    )
    report = process_phase(raw_orders, classic_rules, loaded_state)

    assert len(report.valid_syntax) == 0
    assert len(report.valid_semantics) == 0
    assert len(report.syntax_errors) == 1
    assert len(report.semantic_errors) == 0
    assert len(report.resolution_results) == 1


def test_process_move_phase_invalid_semantic(
    loaded_state_factory, classic_rules
):
    raw_orders = {"P1": ["lon-mun"]}
    loaded_state = loaded_state_factory(
        [("U1", "P1", UnitType.ARMY, "lon")],
        game_meta={"turn_code": "1901-S-M"},
    )
    report = process_phase(raw_orders, classic_rules, loaded_state)

    assert len(report.valid_syntax) == 1
    assert len(report.valid_semantics) == 0
    assert len(report.syntax_errors) == 0
    assert len(report.semantic_errors) == 1
    assert len(report.resolution_results) == 1


def test_retreat_phase_bounce_real(loaded_state_factory, classic_rules):
    unit_specs = [
        ("U1", "P1", UnitType.ARMY, "bel"),
        ("U2", "P2", UnitType.ARMY, "ruh"),
        ("U3", "P2", UnitType.ARMY, "pic"),
        ("U4", "P1", UnitType.ARMY, "bur"),
        ("U5", "P1", UnitType.ARMY, "mun"),
        ("U6", "P2", UnitType.FLEET, "nth"),
    ]

    move_state = loaded_state_factory(
        unit_specs,
        game_meta={"turn_code": "1901-S-M"},
    )

    raw_orders = {
        "P2": ["pic-bel", "ruh hold", "nth s pic - bel"],
        "P1": ["bur-ruh", "bel hold", "mun s bur - ruh"],
    }

    move_report = process_phase(raw_orders, classic_rules, move_state)

    retreat_state = loaded_state_factory(
        unit_specs=unit_specs,
        game_meta={"turn_code": "1901-S-R"},
        pending_move=move_report,
    )

    retreat_orders = {
        "P1": ["bel-hol"],
        "P2": ["ruh-hol"],
    }

    retreat_report = process_phase(retreat_orders, classic_rules, retreat_state)

    assert len(retreat_report.resolution_results) == 2
    outcomes = {r.unit_id: r.outcome for r in retreat_report.resolution_results}
    assert outcomes["U1"] == OutcomeType.RETREAT_FAILED
    assert outcomes["U2"] == OutcomeType.RETREAT_FAILED


def test_adjustment_phase_single_disband(loaded_state_factory, classic_rules):
    loaded_state = loaded_state_factory(
        unit_specs=[("U1", "ger", UnitType.ARMY, "mun")],
        game_meta={"turn_code": "1901-F-A"},
        territory_state={
            "mun": {"owner_id": "ger", "supply_center": True},
        },
    )

    raw_orders = {"ger": ["disband army mun"]}

    report = process_phase(raw_orders, classic_rules, loaded_state)

    assert len(report.resolution_results) == 1
    result = report.resolution_results[0]
    assert result.unit_id == "U1"
    assert result.outcome == OutcomeType.DISBAND_SUCCESS


def test_adjustment_phase_disband_duplicate(
    loaded_state_factory, classic_rules
):
    loaded_state = loaded_state_factory(
        unit_specs=[("U1", "ger", UnitType.ARMY, "mun")],
        game_meta={"turn_code": "1901-F-A"},
        territory_state={"mun": {"owner_id": "ger", "supply_center": True}},
    )

    raw_orders = {"ger": ["disband army mun", "disband army mun"]}

    report = process_phase(raw_orders, classic_rules, loaded_state)

    result = report.resolution_results[0]
    assert result.unit_id == "U1"
    assert result.outcome == OutcomeType.DISBAND_SUCCESS
    assert len(result.duplicate_orders) == 1


def test_adjustment_phase_build_over_cap(loaded_state_factory, classic_rules):
    state = loaded_state_factory(
        unit_specs=[("U1", "ger", UnitType.ARMY, "mun")],
        players={"ger": {"status": "active", "nation_id": "ger"}},
        game_meta={"turn_code": "1901-F-A"},
        territory_state={"ber": {"owner_id": "ger", "supply_center": True}},
    )

    raw_orders = {"ger": ["build army ber"]}

    report = process_phase(raw_orders, classic_rules, state)

    result = report.resolution_results[0]
    assert result.origin_territory == "ber"
    assert result.outcome == OutcomeType.BUILD_NO_CENTER


def test_adjustment_phase_duplicate_build(loaded_state_factory, classic_rules):
    state = loaded_state_factory(
        unit_specs=[],
        players={"ger": {"status": "active", "nation_id": "ger"}},
        game_meta={"turn_code": "1901-F-A"},
        territory_state={"ber": {"owner_id": "ger", "supply_center": True}},
    )

    raw_orders = {"ger": ["build army ber", "build fleet ber"]}

    report = process_phase(raw_orders, classic_rules, state)

    result = report.resolution_results[0]
    assert result.outcome in {
        OutcomeType.BUILD_SUCCESS,
        OutcomeType.BUILD_NO_CENTER,
    }
    assert len(result.duplicate_orders) == 1
