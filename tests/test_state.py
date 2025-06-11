import pytest
from diplomacy_cli.core.logic.state import (
    load_orders,
    load_phase_resolution_report,
    save_phase_resolution_report,
)
from diplomacy_cli.core.logic.schema import PhaseResolutionReport, Season, Phase
import json
from pathlib import Path

from collections import defaultdict

from diplomacy_cli.core.logic.turn_code import format_turn_code


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

    save_phase_resolution_report(game_id, report, save_dir=tmp_path)

    loaded = load_phase_resolution_report(
        game_id, 1901, Season.SPRING, Phase.MOVEMENT, save_dir=tmp_path
    )

    assert loaded == report


def test_load_phase_resolution_report_raises_for_missing_file(tmp_path):
    with pytest.raises(FileNotFoundError):
        load_phase_resolution_report(
            game_id="missing_game",
            year=1901,
            season=Season.SPRING,
            phase=Phase.MOVEMENT,
            save_dir=tmp_path,
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
            save_dir=tmp_path,
        )
