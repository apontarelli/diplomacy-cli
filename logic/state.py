from .storage import load, save, DEFAULT_SAVES_DIR
from .turn_code import INITIAL_TURN_CODE
import shutil, os

def start_game(variant = "classic", game_id = "new_game"):
    print("Starting new game")
    start_path = f"data/{variant}/start"
    save_path = f"{DEFAULT_SAVES_DIR}/{game_id}"

    os.makedirs(save_path, exist_ok=True)
    for file in ["players.json", "units.json", "territory_state.json"]:
        shutil.copyfile(f"{start_path}/{file}", f"{save_path}/{file}")
    
    game = {
        "game_id": game_id,
        "variant": variant,
        "turn_code": INITIAL_TURN_CODE,
        "status": "active"
    }
    save(game, f"{save_path}/game.json")

def load_state(game_id):
    path = f"{DEFAULT_SAVES_DIR}/{game_id}"
    return {
        "game": load(f"{path}/game.json"),
        "players": load(f"{path}/players.json"),
        "territory_state": load(f"{path}/territory_state.json"),
        "units": load(f"{path}/units.json")
    }

def apply_unit_movements(units, movements):
    old_units = units
    new_units = {}
    moved_from = set()

    for move in movements:
        from_territory = move["from"]
        to_territory = move["to"]
        unit = old_units[from_territory]

        new_units[to_territory] = unit
        moved_from.add(from_territory)

    for territory, unit in old_units.items():
        if territory not in moved_from:
            new_units[territory] = unit

    return new_units

def disband_unit(units, territory_id):
    units.pop(territory_id, None)
    return units

def build_unit(units, territory_id, type, owner_id):
    units[territory_id] = {
            "type": type,
            "owner_id": owner_id
            }
    return units

def set_territory_owner(territory_state, territory_id, owner_id):
    territory_state[territory_id] = {
            "owner_id": owner_id
            }
    return territory_state

def eliminate_player(players, player_id):
    players[player_id] = {
            "status": "eliminated"
            }
    return players
