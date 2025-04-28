from .storage import load, save, DEFAULT_SAVES_DIR
from .turn_code import INITIAL_TURN_CODE
import shutil, os

def start_game(variant = "classic", game_id = "new_game"):
    print("Starting new {variant} game")
    variant_path=f"data/{variant}"
    start_path = f"data/{variant}/start"
    save_path = f"{DEFAULT_SAVES_DIR}/{game_id}"

    starting_units = load(f"{start_path}/starting_units.json")
    starting_ownerships = load(f"{start_path}/starting_ownerships.json")
    starting_players = load(f"{start_path}/starting_players.json")

    state = {
        "players": {},
        "units": {},
        "territory_state": {},
        "orders": {},
        "game": {
            "game_id": game_id,
            "variant": variant,
            "turn_code": INITIAL_TURN_CODE,
            "status": "active"
            }
    }

    for player in starting_players:
        state["players"][player["nation_id"]] = { "status": player["status"] }
        
    for o in starting_ownerships:
        territory_id = o["territory_id"]
        owner_id = o["owner_id"] 
        state["territory_state"] = set_territory_owner(
            state["territory_state"],
            territory_id,
            owner_id
        ) 

    counters = {}
    territory_to_unit = {}

    for u in starting_units:
        state["units"], territory_to_unit, counters = build_unit(
            state["units"],
            territory_to_unit,
            counters,
            u["owner_id"],
            u["unit_type"],
            u["location_id"]
        ) 

    if os.path.exists(save_path):
        raise FileExistsError(f"Save directory '{save_path}' already exists.")
    os.makedirs(save_path)
    save(state["players"], f"{save_path}/players.json")  
    save(state["units"], f"{save_path}/units.json")  
    save(state["territory_state"], f"{save_path}/territory_state.json")  
    save(state["game"], f"{save_path}/game.json")
    save(state["orders"], f"{save_path}/orders.json")

    print(f"Game {game_id} created successfully!")

def load_state(game_id):
    path = f"{DEFAULT_SAVES_DIR}/{game_id}"

    state = {
        "game": load(f"{path}/game.json"),
        "players": load(f"{path}/players.json"),
        "territory_state": load(f"{path}/territory_state.json"),
        "units": load(f"{path}/units.json"),
        "orders": load(f"{path}/orders.json"),
            }

    territory_to_unit = build_territory_to_unit(state["units"])
    counters = build_counters(state["units"])

    return state, territory_to_unit, counters

def build_territory_to_unit(units):
    territory_to_unit = {}
    for unit_id, unit in units.items():
        territory_id = unit["territory_id"]
        territory_to_unit[territory_id] = unit_id
    return territory_to_unit

def build_counters(units):
    counters = {}
    for unit_id in units.keys():
        parts = unit_id.split("_")
        if len(parts) != 3:
            continue
        owner, unit_type, number = parts
        key = f"{owner}_{unit_type}"
        counters[key] = max(counters.get(key, 0), int(number))
    return counters

def apply_unit_movements(units, territory_to_unit, movements):
    for move in movements:
        from_ = move["from"]
        to = move["to"]
        unit_id = territory_to_unit[from_]
        units[unit_id]["territory_id"] = to
        territory_to_unit[to] = territory_to_unit.pop(from_)

    return units, territory_to_unit

def disband_unit(units, territory_to_unit, territory_id):
    unit_id = territory_to_unit.pop(territory_id)
    units.pop(unit_id)

    return units, territory_to_unit

def build_unit(units, territory_to_unit, counters, territory_id, unit_type, owner_id):
    key = f"{owner_id}_{unit_type}"
    next_num = counters.get(key, 0) + 1
    unit_id = f"{key}_{next_num}"

    units[unit_id] = {
            "unit_type": unit_type,
            "owner_id": owner_id,
            "territory_id": territory_id
            }
    territory_to_unit[territory_id] = unit_id
    counters[key] = next_num

    return units, territory_to_unit, counters

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
