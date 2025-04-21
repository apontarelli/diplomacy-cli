from .storage import load, save, DEFAULT_SAVES_DIR
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
        "turn": 1901,
        "season": "Spring",
        "phase": "Orders",
        "status": "active"
    }
    save(game, f"{save_path}/game.json")
