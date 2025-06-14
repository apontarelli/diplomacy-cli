import pytest
from diplomacy_cli.core.logic.state import (
    apply_state_mutations,
    load_orders,
    load_phase_resolution_report,
    process_turn,
    save_phase_resolution_report,
)
from diplomacy_cli.core.logic.schema import (
    PhaseResolutionReport,
    Season,
    Phase,
    UnitType,
)
import json

from collections import defaultdict

from diplomacy_cli.core.logic.storage import save
from diplomacy_cli.core.logic.turn_code import format_turn_code
from diplomacy_cli.core.logic.validator.orchestrator import process_phase


def test_load_orders_returns_defaultdict_list():
    result = load_orders()

    assert isinstance(result, defaultdict)
    assert result.default_factory is list
    assert dict(result) == {}


def test_load_orders_supports_appending_orders():
    orders = load_orders()
    orders["ENG"].append("ENG - NTH")
    orders["ENG"].append("ENG - NWG")

    assert orders["ENG"] == ["ENG - NTH", "ENG - NWG"]


def make_sample_report() -> PhaseResolutionReport:
    return PhaseResolutionReport(
        phase=Phase.MOVEMENT,
        season=Season.SPRING,
        year=1901,
        valid_syntax=[],
        valid_semantics=[],
        syntax_errors=[],
        semantic_errors=[],
        resolution_results=[],
    )


def test_save_and_load_phase_resolution_report(tmp_path):
    report = make_sample_report()
    game_id = "test_game"

    save_phase_resolution_report(game_id, report, tmp_path)

    loaded = load_phase_resolution_report(
        game_id, 1901, Season.SPRING, Phase.MOVEMENT, tmp_path
    )

    assert loaded == report


def test_load_phase_resolution_report_raises_for_missing_file(tmp_path):
    game_id = "test_game"
    with pytest.raises(FileNotFoundError):
        load_phase_resolution_report(
            game_id=game_id,
            year=1901,
            season=Season.SPRING,
            phase=Phase.MOVEMENT,
            root_dir=tmp_path,
        )


def test_load_phase_resolution_report_raises_on_invalid_json(tmp_path):
    game_id = "bad_game"
    year = 1901
    season = Season.SPRING
    phase = Phase.MOVEMENT

    report_path = tmp_path / game_id / "reports"
    report_path.mkdir(parents=True)

    filename = f"{format_turn_code(year, season, phase)}_report.json"
    bad_file = report_path / filename
    bad_file.write_text("{ this is not valid json")

    with pytest.raises(json.JSONDecodeError):
        load_phase_resolution_report(
            game_id=game_id,
            year=year,
            season=season,
            phase=phase,
            root_dir=tmp_path,
        )


def test_apply_state_mutations_retreat_disbands_failed(
    loaded_state_factory, classic_rules
):
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
        raw_orders={
            "P2": ["pic-bel", "ruh hold", "nth s pic - bel"],
            "P1": ["bur-ruh", "bel hold", "mun s bur - ruh"],
        },
    )

    move_report = process_phase(move_state, classic_rules)

    retreat_state = loaded_state_factory(
        unit_specs=unit_specs,
        game_meta={"turn_code": "1901-S-R"},
        pending_move=move_report,
        raw_orders={
            "P1": ["bel-hol"],
            "P2": ["ruh-hol"],
        },
    )

    retreat_report = process_phase(retreat_state, classic_rules)

    new_state = apply_state_mutations(retreat_state, retreat_report)

    assert "U1" not in new_state.game.units
    assert "U2" not in new_state.game.units


def test_apply_state_mutations_retreat_successful_movement(
    loaded_state_factory, classic_rules
):
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
        raw_orders={
            "P2": ["pic-bel", "ruh hold", "nth s pic - bel"],
            "P1": ["bur-ruh", "bel hold", "mun s bur - ruh"],
        },
    )

    move_report = process_phase(move_state, classic_rules)

    retreat_state = loaded_state_factory(
        unit_specs=unit_specs,
        game_meta={"turn_code": "1901-S-R"},
        pending_move=move_report,
        raw_orders={
            "P1": ["bel-hol"],
            "P2": ["ruh-kie"],
        },
    )

    retreat_report = process_phase(retreat_state, classic_rules)

    new_state = apply_state_mutations(retreat_state, retreat_report)

    assert "U1" in new_state.game.units
    assert "U2" in new_state.game.units
    assert new_state.territory_to_unit["hol"] == "U1"
    assert new_state.territory_to_unit["kie"] == "U2"


def test_apply_state_mutations_adjustment_disband_and_build(
    loaded_state_factory, classic_rules
):
    unit_specs = [
        ("fra_army_1", "fra", UnitType.ARMY, "bel"),
        ("fra_army_2", "fra", UnitType.ARMY, "pic"),
    ]

    adjustment_state = loaded_state_factory(
        unit_specs=unit_specs,
        game_meta={"turn_code": "1901-W-A"},
        pending_move=None,
        territory_state={
            "par": {"owner_id": "fra", "supply_center": True},
            "ber": {"owner_id": "fra", "supply_center": True},
        },
        raw_orders={
            "fra": ["disband army bel", "build army par"],
        },
    )

    adjustment_report = process_phase(adjustment_state, classic_rules)
    print(adjustment_report)
    new_state = apply_state_mutations(adjustment_state, adjustment_report)

    assert "fra_army_3" in new_state.game.units.keys()
    assert "fra_army_2" in new_state.game.units.keys()
    assert "par" in new_state.territory_to_unit.keys()
    assert "pic" in new_state.territory_to_unit.keys()
    assert "bel" not in new_state.territory_to_unit.keys()


def test_process_turn_advances_phase_and_mutates_state(
    tmp_path, loaded_state_factory
):
    game_id = "test_game"

    unit_specs = [
        ("U1", "P1", UnitType.ARMY, "par"),
        ("U2", "P2", UnitType.ARMY, "bur"),
    ]

    initial_state = loaded_state_factory(
        unit_specs=unit_specs,
        game_meta={"turn_code": "1901-S-M", "variant": "classic"},
        raw_orders={
            "P1": ["par-bur"],
            "P2": ["bur hold"],
        },
    )

    save_root = tmp_path / game_id
    save_root.mkdir(parents=True)

    save(initial_state.game.players, save_root / "players.json")
    save(initial_state.game.units, save_root / "units.json")
    save(initial_state.game.territory_state, save_root / "territory_state.json")
    save(initial_state.game.game_meta, save_root / "game.json")
    save(initial_state.game.raw_orders, save_root / "orders.json")

    new_state = process_turn(game_id, root_dir=tmp_path)

    assert new_state.game.game_meta["turn_code"] == "1901-F-M"

    assert isinstance(new_state.game.units, dict)
    assert new_state.game.units

    assert new_state.game.raw_orders == {}

    report_file = save_root / "reports" / "1901-S-M_report.json"
    assert report_file.exists()

    with report_file.open() as f:
        report_data = json.load(f)
        assert "resolution_results" in report_data
