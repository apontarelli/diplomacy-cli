import pytest

from diplomacy_cli.core.logic.schema import (
    Order,
    OrderType,
    OutcomeType,
    Phase,
    PhaseResolutionReport,
    ResolutionResult,
    Season,
    SemanticResult,
    SyntaxResult,
    UnitType,
)
from diplomacy_cli.core.logic.validator.semantic import (
    SemanticError,
    _check_adjacency,
    _check_build,
    _check_convoy,
    _check_disband,
    _check_hold,
    _check_move,
    _check_retreat,
    _check_support_hold,
    _check_support_move,
    _check_territory_exists,
    _check_unit_exists,
    _check_unit_ownership,
    _has_sea_path,
    validate_semantic,
)


def test_territory_exists(classic_rules):
    _check_territory_exists("lon", classic_rules.territory_ids)
    _check_territory_exists("stp_sc", classic_rules.territory_ids)


def test_territory_exists_invalid(classic_rules):
    with pytest.raises(SemanticError, match="fun is not a valid territory"):
        _check_territory_exists("fun", classic_rules.territory_ids)


def test_check_unit_exists_valid(loaded_state_factory):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    _check_unit_exists("lon", state.territory_to_unit)


def test_check_unit_exists_invalid(loaded_state_factory):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    order = Order(origin="par", order_type=OrderType.MOVE, destination="wal")
    with pytest.raises(SemanticError, match="Unit does not exist in par"):
        _check_unit_exists(order.origin, state.territory_to_unit)


def test_check_unit_ownership_valid(loaded_state_factory):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    _check_unit_ownership(
        "eng",
        "lon",
        state.game.units,
        state.territory_to_unit,
    )


def test_check_unit_ownership_invalid(loaded_state_factory):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    with pytest.raises(
        SemanticError, match="Unit in lon does not belong to fra"
    ):
        _check_unit_ownership(
            "fra",
            "lon",
            state.game.units,
            state.territory_to_unit,
        )


def test_check_adjacency_valid(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    _check_adjacency("lon", "wal", state, classic_rules)


def test_check_adjacency_valid_sea(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    _check_adjacency("lon", "pic", state, classic_rules, allow_convoy=True)


def test_check_adjacency_invalid_territory(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    with pytest.raises(SemanticError, match="Army at lon cannot reach mos"):
        _check_adjacency("lon", "mos", state, classic_rules)


def test_check_adjacency_invalid_unit_type(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    with pytest.raises(SemanticError, match="Army at lon cannot reach eng"):
        _check_adjacency("lon", "eng", state, classic_rules)


def test_check_adjacency_fleet_valid(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    state = loaded_state_factory(unit_specs)
    _check_adjacency("bre", "eng", state, classic_rules)


def test_check_adjacency_fleet_invalid(loaded_state_factory, classic_rules):
    unit_specs = [
        ("U1", "eng", UnitType.FLEET, "bre"),
        ("U2", "eng", UnitType.FLEET, "stp_sc"),
    ]
    state = loaded_state_factory(unit_specs)
    with pytest.raises(SemanticError, match="Fleet at bre cannot reach par"):
        _check_adjacency("bre", "par", state, classic_rules)
    with pytest.raises(SemanticError, match="Fleet at stp_sc cannot reach bar"):
        _check_adjacency("stp_sc", "bar", state, classic_rules)


@pytest.mark.parametrize(
    "origin,target,expected",
    [
        ("lon", "bre", True),  # London → Brest via NTH → ENG → BRE
        ("spa", "bre", True),  # Spain → Brest via MAO
        ("mun", "ber", False),  # Munich → Berlin: both landlocked
        ("stp", "smy", True),  # Very long convoy
    ],
)
def test_has_sea_path_various_land_pairs(
    classic_rules, origin, target, expected
):
    assert _has_sea_path(origin, target, classic_rules) is expected


def test_symmetry(classic_rules):
    """Sea-path test should be symmetric."""
    assert _has_sea_path("lon", "bre", classic_rules) == _has_sea_path(
        "bre", "lon", classic_rules
    )


def test_check_build(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    territory_state = {}
    territory_state["edi"] = {"owner_id": "eng"}
    state = loaded_state_factory(unit_specs, territory_state)
    build_order = Order(
        origin="edi", order_type=OrderType.BUILD, unit_type=UnitType.ARMY
    )
    player_id = "eng"
    _check_build(player_id, build_order, state, classic_rules)


def test_check_build_occupied(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "lon")]
    territory_state = {}
    territory_state["lon"] = {"owner_id": "eng"}
    state = loaded_state_factory(unit_specs, territory_state)
    build_order = Order(
        origin="lon", order_type=OrderType.BUILD, unit_type=UnitType.ARMY
    )
    player_id = "eng"
    with pytest.raises(
        SemanticError, match="Cannot build in lon: territory is occupied"
    ):
        _check_build(player_id, build_order, state, classic_rules)


def test_check_build_not_home(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    territory_state = {}
    territory_state["mar"] = {"owner_id": "eng"}
    state = loaded_state_factory(unit_specs, territory_state)
    build_order = Order(
        origin="mar", order_type=OrderType.BUILD, unit_type=UnitType.ARMY
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="mar is not a home center of eng"):
        _check_build(player_id, build_order, state, classic_rules)


def test_check_build_not_owned(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    territory_state = {}
    territory_state["lvp"] = {"owner_id": "fra"}
    state = loaded_state_factory(unit_specs, territory_state)
    build_order = Order(
        origin="lvp", order_type=OrderType.BUILD, unit_type=UnitType.ARMY
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="lvp does not belong to eng"):
        _check_build(player_id, build_order, state, classic_rules)


def test_check_build_land_fleet(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    territory_state = {}
    territory_state["par"] = {"owner_id": "fra"}
    state = loaded_state_factory(unit_specs, territory_state)
    build_order = Order(
        origin="par", order_type=OrderType.BUILD, unit_type=UnitType.FLEET
    )
    player_id = "fra"
    with pytest.raises(
        SemanticError, match="Fleets can only be built on coasts"
    ):
        _check_build(player_id, build_order, state, classic_rules)


def test_check_disband(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    disband_order = Order(
        origin="lon", order_type=OrderType.DISBAND, unit_type=UnitType.ARMY
    )
    player_id = "eng"
    _check_disband(player_id, disband_order, state, classic_rules)


def test_check_disband_no_unit(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    state = loaded_state_factory(unit_specs)
    disband_order = Order(
        origin="swe", order_type=OrderType.DISBAND, unit_type=UnitType.FLEET
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="Unit does not exist in swe"):
        _check_disband(
            player_id,
            disband_order,
            state,
            classic_rules,
        )


def test_check_disband_wrong_unit(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    disband_order = Order(
        origin="lon", order_type=OrderType.DISBAND, unit_type=UnitType.FLEET
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="No fleet at lon"):
        _check_disband(
            player_id,
            disband_order,
            state,
            classic_rules,
        )


def test_check_disband_not_owned(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "fra", UnitType.ARMY, "pie")]
    state = loaded_state_factory(unit_specs)
    disband_order = Order(
        origin="pie", order_type=OrderType.DISBAND, unit_type=UnitType.ARMY
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="pie does not belong to eng"):
        _check_disband(
            player_id,
            disband_order,
            state,
            classic_rules,
        )


def test_check_support_move(loaded_state_factory, classic_rules):
    unit_specs = [
        ("U1", "eng", UnitType.FLEET, "eng"),
        ("U2", "eng", UnitType.ARMY, "lon"),
    ]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="eng",
        order_type=OrderType.SUPPORT_MOVE,
        support_origin="lon",
        support_destination="wal",
    )
    _check_support_move("eng", order, state, classic_rules)


def test_check_support_hold(loaded_state_factory, classic_rules):
    unit_specs = [
        ("U1", "eng", UnitType.FLEET, "eng"),
        ("U2", "eng", UnitType.ARMY, "lon"),
    ]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="eng", order_type=OrderType.SUPPORT_HOLD, support_origin="lon"
    )
    _check_support_hold("eng", order, state, classic_rules)


def test_check_convoy(loaded_state_factory, classic_rules):
    unit_specs = [
        ("U1", "eng", UnitType.FLEET, "eng"),
        ("U2", "eng", UnitType.ARMY, "lon"),
    ]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="eng",
        order_type=OrderType.CONVOY,
        convoy_origin="lon",
        convoy_destination="pic",
    )
    _check_convoy("eng", order, state, classic_rules)


def test_check_convoy_missing_dest(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="eng", order_type=OrderType.CONVOY, convoy_origin="lon"
    )
    with pytest.raises(
        SemanticError, match="Convoy must specify both origin and destination"
    ):
        _check_convoy("eng", order, state, classic_rules)


def test_check_convoy_army(loaded_state_factory, classic_rules):
    unit_specs = [
        ("U1", "eng", UnitType.ARMY, "eng"),
        ("U2", "eng", UnitType.ARMY, "lon"),
    ]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="lon",
        order_type=OrderType.CONVOY,
        convoy_origin="eng",
        convoy_destination="pic",
    )
    with pytest.raises(SemanticError, match="No fleet at lon to convoy"):
        _check_convoy("eng", order, state, classic_rules)


def test_check_convoy_fleet(loaded_state_factory, classic_rules):
    unit_specs = [
        ("U1", "eng", UnitType.FLEET, "eng"),
        ("U2", "eng", UnitType.FLEET, "bre"),
    ]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="eng",
        order_type=OrderType.CONVOY,
        convoy_origin="bre",
        convoy_destination="wal",
    )
    with pytest.raises(SemanticError, match="No army at bre to convoy"):
        _check_convoy("eng", order, state, classic_rules)


def test_check_convoy_no_path(loaded_state_factory, classic_rules):
    unit_specs = [
        ("U1", "eng", UnitType.FLEET, "eng"),
        ("U2", "eng", UnitType.ARMY, "lon"),
    ]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="eng",
        order_type=OrderType.CONVOY,
        convoy_origin="lon",
        convoy_destination="mun",
    )
    with pytest.raises(
        SemanticError, match="No valid sea path between lon and mun"
    ):
        _check_convoy("eng", order, state, classic_rules)


def test_check_move(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "lon")]
    state = loaded_state_factory(unit_specs)
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="wal")
    _check_move("eng", order, state, classic_rules)


def test_check_invalid(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs)
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="mun")
    with pytest.raises(
        SemanticError,
        match="Army at lon cannot reach mun: "
        "no continuous sea route for convoy",
    ):
        _check_move("eng", order, state, classic_rules)


def test_check_hold(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "lon")]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="lon",
        order_type=OrderType.HOLD,
    )
    _check_hold("eng", order, state, classic_rules)


def test_check_invalid_hold(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    state = loaded_state_factory(unit_specs)
    order = Order(
        origin="wal",
        order_type=OrderType.HOLD,
    )
    with pytest.raises(SemanticError):
        _check_hold("eng", order, state, classic_rules)


def test_check_retreat(loaded_state_factory, classic_rules):
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="wal")
    semantic = SemanticResult("eng", "lon-wal", "lon-wal", order, True, [])

    resolution = ResolutionResult(
        unit_id="england_army_1",
        owner_id="eng",
        unit_type=UnitType.ARMY,
        origin_territory="lon",
        semantic_result=semantic,
        outcome=OutcomeType.DISLODGED,
        resolved_territory="lon",
        strength=1,
        dislodged_by_id=None,
        destination="wal",
        convoy_path=None,
        supported_unit_id=None,
        duplicate_orders=[],
    )

    report = PhaseResolutionReport(
        phase=Phase.MOVEMENT,
        season=Season.SPRING,
        year=1901,
        valid_syntax=[],
        valid_semantics=[],
        syntax_errors=[],
        semantic_errors=[],
        resolution_results=[resolution],
    )

    unit_specs = [("U1", "eng", UnitType.FLEET, "lon")]
    state = loaded_state_factory(unit_specs=unit_specs, pending_move=report)
    player_id = "eng"
    order = Order(origin="lon", order_type=OrderType.RETREAT, destination="wal")
    _check_retreat(player_id, order, state, classic_rules)


def test_check_retreat_no_destination(loaded_state_factory, classic_rules):
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="wal")
    semantic = SemanticResult("eng", "lon-wal", "lon-wal", order, True, [])

    resolution = ResolutionResult(
        unit_id="england_army_1",
        owner_id="eng",
        unit_type=UnitType.ARMY,
        origin_territory="lon",
        semantic_result=semantic,
        outcome=OutcomeType.DISLODGED,
        resolved_territory="lon",
        strength=1,
        dislodged_by_id=None,
        destination="wal",
        convoy_path=None,
        supported_unit_id=None,
        duplicate_orders=[],
    )

    report = PhaseResolutionReport(
        phase=Phase.MOVEMENT,
        season=Season.SPRING,
        year=1901,
        valid_syntax=[],
        valid_semantics=[],
        syntax_errors=[],
        semantic_errors=[],
        resolution_results=[resolution],
    )
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    state = loaded_state_factory(unit_specs=unit_specs, pending_move=report)
    player_id = "eng"
    order = Order(origin="lon", order_type=OrderType.RETREAT)
    with pytest.raises(
        SemanticError, match="Retreat must specify a destination"
    ):
        _check_retreat(player_id, order, state, classic_rules)


def test_check_retreat_not_dislodged(loaded_state_factory, classic_rules):
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="wal")
    semantic = SemanticResult("eng", "lon-wal", "lon-wal", order, True, [])

    resolution = ResolutionResult(
        unit_id="england_army_1",
        owner_id="eng",
        unit_type=UnitType.ARMY,
        origin_territory="lon",
        semantic_result=semantic,
        outcome=OutcomeType.DISLODGED,
        resolved_territory="lon",
        strength=1,
        dislodged_by_id=None,
        destination="wal",
        convoy_path=None,
        supported_unit_id=None,
        duplicate_orders=[],
    )

    report = PhaseResolutionReport(
        phase=Phase.MOVEMENT,
        season=Season.SPRING,
        year=1901,
        valid_syntax=[],
        valid_semantics=[],
        syntax_errors=[],
        semantic_errors=[],
        resolution_results=[resolution],
    )
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    state = loaded_state_factory(unit_specs=unit_specs, pending_move=report)
    player_id = "eng"
    order = Order(origin="eng", order_type=OrderType.RETREAT, destination="iri")
    with pytest.raises(SemanticError, match="No dislodged unit at eng"):
        _check_retreat(player_id, order, state, classic_rules)


def test_check_retreat_occupied(loaded_state_factory, classic_rules):
    order = Order(origin="eng", order_type=OrderType.MOVE, destination="wal")
    semantic = SemanticResult("eng", "eng-wal", "eng-wal", order, True, [])

    resolution = ResolutionResult(
        unit_id="england_army_1",
        owner_id="eng",
        unit_type=UnitType.ARMY,
        origin_territory="eng",
        semantic_result=semantic,
        outcome=OutcomeType.DISLODGED,
        resolved_territory="eng",
        strength=1,
        dislodged_by_id=None,
        destination="wal",
        convoy_path=None,
        supported_unit_id=None,
        duplicate_orders=[],
    )

    report = PhaseResolutionReport(
        phase=Phase.MOVEMENT,
        season=Season.SPRING,
        year=1901,
        valid_syntax=[],
        valid_semantics=[],
        syntax_errors=[],
        semantic_errors=[],
        resolution_results=[resolution],
    )
    unit_specs = [
        ("U1", "eng", UnitType.FLEET, "eng"),
        ("U2", "eng", UnitType.FLEET, "bre"),
    ]
    state = loaded_state_factory(unit_specs=unit_specs, pending_move=report)
    player_id = "eng"
    order = Order(origin="eng", order_type=OrderType.RETREAT, destination="bre")
    with pytest.raises(SemanticError, match="bre is occupied"):
        _check_retreat(player_id, order, state, classic_rules)


def test_validate_semantic(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.ARMY, "lon")]
    state = loaded_state_factory(unit_specs=unit_specs)
    player_id = "eng"
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="wal")
    syntax = SyntaxResult(
        player_id=player_id,
        raw="lon - wal",
        normalized="lon-wal",
        valid=True,
        errors=[],
        order=order,
    )

    expected = SemanticResult(
        player_id="eng",
        raw="lon - wal",
        normalized="lon-wal",
        valid=True,
        errors=[],
        order=order,
    )
    validated = validate_semantic(player_id, syntax, classic_rules, state)

    assert validated == expected


def test_validate_semantic_invalid(loaded_state_factory, classic_rules):
    unit_specs = [("U1", "eng", UnitType.FLEET, "bre")]
    state = loaded_state_factory(unit_specs=unit_specs)
    player_id = "eng"
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="fun")
    syntax = SyntaxResult(
        player_id=player_id,
        raw="lon - fun",
        normalized="lon-fun",
        valid=True,
        errors=[],
        order=order,
    )

    expected = SemanticResult(
        player_id="eng",
        raw="lon - fun",
        normalized="lon-fun",
        valid=False,
        errors=["fun is not a valid territory"],
        order=order,
    )
    validated = validate_semantic(player_id, syntax, classic_rules, state)

    assert validated == expected
