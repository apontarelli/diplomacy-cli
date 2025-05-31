import pytest

from diplomacy_cli.core.logic.rules_loader import load_rules
from diplomacy_cli.core.logic.schema import (
    GameState,
    LoadedState,
    Order,
    OrderType,
    SemanticResult,
    SyntaxResult,
)
from diplomacy_cli.core.logic.state import (
    build_counters,
    build_territory_to_unit,
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


@pytest.fixture
def classic_rules():
    return load_rules("classic")


@pytest.fixture
def loaded_state():
    territory_state = {
        "lon": {"territory_id": "lon", "owner_id": "eng"},
        "edi": {"territory_id": "edi", "owner_id": "eng"},
        "bre": {"territory_id": "bre", "owner_id": "fra"},
        "mar": {"territory_id": "bre", "owner_id": "fra"},
        "swe": {"territory_id": "swe", "owner_id": "eng"},
        "lvp": {"territory_id": "lvp", "owner_id": "rus"},
        "par": {"territory_id": "par", "owner_id": "fra"},
    }

    units = {
        "england_army_1": {
            "id": "england_army_1",
            "unit_type": "army",
            "owner_id": "eng",
            "territory_id": "lon",
        },
        "england_fleet_1": {
            "id": "england_fleet_1",
            "unit_type": "fleet",
            "owner_id": "eng",
            "territory_id": "eng",
        },
        "france_army_1": {
            "id": "france_army_1",
            "unit_type": "army",
            "owner_id": "fra",
            "territory_id": "pie",
        },
        "france_fleet_1": {
            "id": "france_fleet_1",
            "unit_type": "fleet",
            "owner_id": "fra",
            "territory_id": "bre",
        },
        "rus_fleet_1": {
            "id": "rus_fleet_1",
            "unit_type": "fleet",
            "owner_id": "rus",
            "territory_id": "stp_sc",
        },
    }

    game = GameState(
        game_meta={
            "game_id": "test_game",
            "variant": "classic",
            "turn_code": "S1901M",
            "dislodged": set(),
        },
        players={
            "england": {"status": "active"},
            "france": {"status": "active"},
        },
        territory_state=territory_state,
        units=units,
        raw_orders={},
    )

    return LoadedState(
        game=game,
        territory_to_unit=build_territory_to_unit(game.units),
        counters=build_counters(game.units),
        dislodged=set(),
    )


@pytest.fixture
def order_london():
    return Order(origin="lon", order_type=OrderType.MOVE, destination="wal")


def test_territory_exists(classic_rules):
    _check_territory_exists("lon", classic_rules.territory_ids)
    _check_territory_exists("stp_sc", classic_rules.territory_ids)


def test_territory_exists_invalid(classic_rules):
    with pytest.raises(SemanticError, match="fun is not a valid territory"):
        _check_territory_exists("fun", classic_rules.territory_ids)


def test_check_unit_exists_valid(order_london, loaded_state):
    _check_unit_exists(order_london.origin, loaded_state.territory_to_unit)


def test_check_unit_exists_invalid(loaded_state):
    order = Order(origin="par", order_type=OrderType.MOVE, destination="wal")
    with pytest.raises(SemanticError, match="Unit does not exist in par"):
        _check_unit_exists(order.origin, loaded_state.territory_to_unit)


def test_check_unit_ownership_valid(order_london, loaded_state):
    _check_unit_ownership(
        "eng",
        order_london.origin,
        loaded_state.game.units,
        loaded_state.territory_to_unit,
    )


def test_check_unit_ownership_invalid(order_london, loaded_state):
    with pytest.raises(
        SemanticError, match="Unit in lon does not belong to france"
    ):
        _check_unit_ownership(
            "france",
            order_london.origin,
            loaded_state.game.units,
            loaded_state.territory_to_unit,
        )


def test_check_adjacency_valid(loaded_state, classic_rules):
    _check_adjacency("lon", "wal", loaded_state, classic_rules)


def test_check_adjacency_valid_sea(loaded_state, classic_rules):
    _check_adjacency(
        "lon", "pic", loaded_state, classic_rules, allow_convoy=True
    )


def test_check_adjacency_invalid_territory(loaded_state, classic_rules):
    with pytest.raises(SemanticError, match="Army at lon cannot reach mos"):
        _check_adjacency("lon", "mos", loaded_state, classic_rules)


def test_check_adjacency_invalid_unit_type(loaded_state, classic_rules):
    with pytest.raises(SemanticError, match="Army at lon cannot reach eng"):
        _check_adjacency("lon", "eng", loaded_state, classic_rules)


def test_check_adjacency_fleet_valid(loaded_state, classic_rules):
    _check_adjacency("bre", "eng", loaded_state, classic_rules)


def test_check_adjacency_fleet_invalid(loaded_state, classic_rules):
    with pytest.raises(SemanticError, match="Fleet at bre cannot reach par"):
        _check_adjacency("bre", "par", loaded_state, classic_rules)
    with pytest.raises(SemanticError, match="Fleet at stp_sc cannot reach bar"):
        _check_adjacency("stp_sc", "bar", loaded_state, classic_rules)


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


def test_check_build(loaded_state, classic_rules):
    build_order = Order(
        origin="edi", order_type=OrderType.BUILD, unit_type="army"
    )
    player_id = "eng"
    _check_build(player_id, build_order, loaded_state, classic_rules)


def test_check_build_occupied(loaded_state, classic_rules):
    build_order = Order(
        origin="lon", order_type=OrderType.BUILD, unit_type="army"
    )
    player_id = "eng"
    with pytest.raises(
        SemanticError, match="Cannot build in lon: territory is occupied"
    ):
        _check_build(player_id, build_order, loaded_state, classic_rules)


def test_check_build_not_home(loaded_state, classic_rules):
    build_order = Order(
        origin="mar", order_type=OrderType.BUILD, unit_type="army"
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="mar is not a home center of eng"):
        _check_build(player_id, build_order, loaded_state, classic_rules)


def test_check_build_supply_centers(loaded_state, classic_rules):
    build_order = Order(
        origin="mar", order_type=OrderType.BUILD, unit_type="army"
    )
    loaded_state.game.units["france_army_2"] = {
        "id": "france_army_2",
        "unit_type": "army",
        "owner_id": "fra",
        "territory_id": "spa",
    }
    player_id = "fra"
    with pytest.raises(
        SemanticError,
        match="fra does not have enough supply centers to build a unit",
    ):
        _check_build(player_id, build_order, loaded_state, classic_rules)


def test_check_build_not_owned(loaded_state, classic_rules):
    build_order = Order(
        origin="lvp", order_type=OrderType.BUILD, unit_type="army"
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="lvp does not belong to eng"):
        _check_build(player_id, build_order, loaded_state, classic_rules)


def test_check_build_land_fleet(loaded_state, classic_rules):
    build_order = Order(
        origin="par", order_type=OrderType.BUILD, unit_type="fleet"
    )
    player_id = "fra"
    with pytest.raises(
        SemanticError, match="Fleets can only be built on coasts"
    ):
        _check_build(player_id, build_order, loaded_state, classic_rules)


def test_check_disband(loaded_state, classic_rules):
    disband_order = Order(
        origin="lon", order_type=OrderType.DISBAND, unit_type="army"
    )
    player_id = "eng"
    _check_disband(player_id, disband_order, loaded_state, classic_rules)


def test_check_disband_no_unit(loaded_state, classic_rules):
    disband_order = Order(
        origin="swe", order_type=OrderType.DISBAND, unit_type="fleet"
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="Unit does not exist in swe"):
        _check_disband(
            player_id,
            disband_order,
            loaded_state,
            classic_rules,
        )


def test_check_disband_wrong_unit(loaded_state, classic_rules):
    disband_order = Order(
        origin="lon", order_type=OrderType.DISBAND, unit_type="fleet"
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="No fleet at lon"):
        _check_disband(
            player_id,
            disband_order,
            loaded_state,
            classic_rules,
        )


def test_check_disband_not_owned(loaded_state, classic_rules):
    disband_order = Order(
        origin="pie", order_type=OrderType.DISBAND, unit_type="army"
    )
    player_id = "eng"
    with pytest.raises(SemanticError, match="pie does not belong to eng"):
        _check_disband(
            player_id,
            disband_order,
            loaded_state,
            classic_rules,
        )


def test_check_support_move(loaded_state, classic_rules):
    order = Order(
        origin="eng",
        order_type=OrderType.SUPPORT_MOVE,
        support_origin="lon",
        support_destination="wal",
    )
    _check_support_move("eng", order, loaded_state, classic_rules)


def test_check_support_hold(loaded_state, classic_rules):
    order = Order(
        origin="eng", order_type=OrderType.SUPPORT_HOLD, support_origin="lon"
    )
    _check_support_hold("eng", order, loaded_state, classic_rules)


def test_check_convoy(loaded_state, classic_rules):
    order = Order(
        origin="eng",
        order_type=OrderType.CONVOY,
        convoy_origin="lon",
        convoy_destination="pic",
    )
    _check_convoy("eng", order, loaded_state, classic_rules)


def test_check_convoy_missing_dest(loaded_state, classic_rules):
    order = Order(
        origin="eng", order_type=OrderType.CONVOY, convoy_origin="lon"
    )
    with pytest.raises(
        SemanticError, match="Convoy must specify both origin and destination"
    ):
        _check_convoy("eng", order, loaded_state, classic_rules)


def test_check_convoy_army(loaded_state, classic_rules):
    order = Order(
        origin="lon",
        order_type=OrderType.CONVOY,
        convoy_origin="eng",
        convoy_destination="pic",
    )
    with pytest.raises(SemanticError, match="No fleet at lon to convoy"):
        _check_convoy("eng", order, loaded_state, classic_rules)


def test_check_convoy_fleet(loaded_state, classic_rules):
    order = Order(
        origin="eng",
        order_type=OrderType.CONVOY,
        convoy_origin="bre",
        convoy_destination="wal",
    )
    with pytest.raises(SemanticError, match="No army at bre to convoy"):
        _check_convoy("eng", order, loaded_state, classic_rules)


def test_check_convoy_no_path(loaded_state, classic_rules):
    order = Order(
        origin="eng",
        order_type=OrderType.CONVOY,
        convoy_origin="lon",
        convoy_destination="mun",
    )
    with pytest.raises(
        SemanticError, match="No valid sea path between lon and mun"
    ):
        _check_convoy("eng", order, loaded_state, classic_rules)


def test_check_move(loaded_state, classic_rules):
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="wal")
    _check_move("eng", order, loaded_state, classic_rules)


def test_check_invalid(loaded_state, classic_rules):
    order = Order(origin="lon", order_type=OrderType.MOVE, destination="mun")
    with pytest.raises(
        SemanticError,
        match="Army at lon cannot reach mun: "
        "no continuous sea route for convoy",
    ):
        _check_move("eng", order, loaded_state, classic_rules)


def test_check_hold(loaded_state, classic_rules):
    order = Order(
        origin="lon",
        order_type=OrderType.HOLD,
    )
    _check_hold("eng", order, loaded_state, classic_rules)


def test_check_invalid_hold(loaded_state, classic_rules):
    order = Order(
        origin="wal",
        order_type=OrderType.HOLD,
    )
    with pytest.raises(SemanticError):
        _check_hold("eng", order, loaded_state, classic_rules)


def test_check_retreat(loaded_state, classic_rules):
    loaded_state.dislodged.add("lon")
    player_id = "eng"
    order = Order(origin="lon", order_type=OrderType.RETREAT, destination="wal")
    _check_retreat(player_id, order, loaded_state, classic_rules)


def test_check_retreat_no_destination(loaded_state, classic_rules):
    loaded_state.dislodged.add("lon")
    player_id = "eng"
    order = Order(origin="lon", order_type=OrderType.RETREAT)
    with pytest.raises(
        SemanticError, match="Retreat must specify a destination"
    ):
        _check_retreat(player_id, order, loaded_state, classic_rules)


def test_check_retreat_not_dislodged(loaded_state, classic_rules):
    loaded_state.dislodged.add("lon")
    player_id = "eng"
    order = Order(origin="eng", order_type=OrderType.RETREAT, destination="iri")
    with pytest.raises(SemanticError, match="No dislodged unit at eng"):
        _check_retreat(player_id, order, loaded_state, classic_rules)


def test_check_retreat_occupied(loaded_state, classic_rules):
    loaded_state.dislodged.add("eng")
    player_id = "eng"
    order = Order(origin="eng", order_type=OrderType.RETREAT, destination="bre")
    with pytest.raises(SemanticError, match="bre is occupied"):
        _check_retreat(player_id, order, loaded_state, classic_rules)


def test_validate_semantic(loaded_state, classic_rules):
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
    validated = validate_semantic(
        player_id, syntax, classic_rules, loaded_state
    )

    assert validated == expected


def test_validate_semantic_invalid(loaded_state, classic_rules):
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
    validated = validate_semantic(
        player_id, syntax, classic_rules, loaded_state
    )

    assert validated == expected
