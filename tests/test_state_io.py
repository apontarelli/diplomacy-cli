import pytest

from diplomacy_cli.core.logic.schema import GameState
from diplomacy_cli.core.logic.state import (
    INITIAL_TURN_CODE,
    build_counters,
    build_territory_to_unit,
    load_state,
    start_game,
)
from diplomacy_cli.core.logic.storage import load


@pytest.fixture
def root(tmp_path):
    return tmp_path


def test_round_trip_identity(root):
    game_id = "io_test"

    new_game = start_game(game_id=game_id, root_dir=root)
    assert isinstance(new_game, GameState)

    base = root / game_id
    expected_files = [
        "players.json",
        "units.json",
        "territory_state.json",
        "game.json",
        "orders.json",
    ]
    for fname in expected_files:
        path = base / fname
        assert path.exists(), f"{fname} missing"
        assert path.stat().st_size > 0, f"{fname} empty"

    loaded = load_state(game_id, root_dir=root)
    saved_players = loaded.game.players
    saved_units = loaded.game.units
    saved_terr = loaded.game.territory_state
    saved_game_meta = loaded.game.game_meta

    assert saved_players == load(base / "players.json")
    assert saved_units == load(base / "units.json")
    assert saved_terr == load(base / "territory_state.json")
    assert saved_game_meta["game_id"] == game_id
    assert saved_game_meta["turn_code"] == INITIAL_TURN_CODE

    assert loaded.territory_to_unit == build_territory_to_unit(saved_units)
    assert loaded.counters == build_counters(saved_units)


def test_start_game_overwrite_protection(root):
    game_id = "clobber_me"

    start_game(game_id=game_id, root_dir=root)

    with pytest.raises(FileExistsError):
        start_game(game_id=game_id, root_dir=root)
