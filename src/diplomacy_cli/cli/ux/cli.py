from diplomacy_cli.core.logic.schema import LoadedState
from diplomacy_cli.core.logic.state import (
    load_state,
    process_turn,
    save_player_orders,
    start_game,
)
from diplomacy_cli.core.paths import delete_game, list_game_ids

from .pretty import format_orders, format_players, format_state


def main_menu():
    while True:
        print("\n--- Diplomacy CLI ---")
        print("1. Start new game")
        print("2. View games")
        print("3. Exit")

        print("--------------------")
        choice = input("Choose an option: ").strip()
        if choice == "1":
            start_new_game()
        elif choice == "2":
            view_games()
        elif choice == "3":
            print("Exiting...")
            break
        else:
            print("Invalid choice.")


def start_new_game():
    print("--------------------")
    game_id = input("Enter a game name: ").strip()
    variant = "classic"
    if game_id:
        start_game(variant=variant, game_id=game_id)
    else:
        print("Game name cannot be blank")


def view_games():
    while True:
        games = list_game_ids()
        if not games:
            print("No saved games")
            return

        print("\n--- Saved games ---")
        for i, game in enumerate(games):
            print(f"{i + 1}. {game}")
        print("0. Back to menu")

        print("--------------------")
        choice = input("Select a save: ").strip()
        try:
            selection = int(choice)
        except ValueError:
            print("Invalid selection.")
            continue

        if selection == 0:
            return

        if not (1 <= selection <= len(games)):
            print("Invalid selection.")
            continue

        selected_game = games[selection - 1]
        manage_save(selected_game)
        return


def manage_save(game_id):
    while True:
        print(f"\n--- {game_id} ---")
        print("1. Print state")
        print("2. Manage orders")
        print("3. Process turn")
        print("4. Delete game")
        print("0. Back to menu")
        print("--------------------")
        choice = input("Choose an option: ").strip()
        if choice == "1":
            state = load_state(game_id)
            print(format_state(state))
        elif choice == "2":
            state = load_state(game_id)
            choose_player(state)
        elif choice == "3":
            process_turn(game_id)
        elif choice == "4":
            confirm = input(
                f"Type exact game_id {game_id} to confirm delete: "
            ).strip()
            if confirm == game_id:
                print("Deleting game")
                delete_game(game_id)
                return
            else:
                print("Input did not match")
        elif choice == "0":
            return
        else:
            print("Invalid choice")


def choose_player(loaded_state: LoadedState):
    while True:
        print(f"\n--- {loaded_state.game.game_meta['turn_code']} ---")
        print(format_players(loaded_state.game.players, "active"))
        print(format_players(loaded_state.game.players, "eliminated"))
        print("0. Go back")
        choice = input("Choose an player: ").strip()
        if choice in loaded_state.game.players:
            manage_orders(loaded_state, choice)
        elif choice == "0":
            return
        else:
            print("Invalid choice")


def manage_orders(loaded_state: LoadedState, player: str):
    players = loaded_state.game.players
    if player not in players.keys():
        print(f"No such player: {player}")

    status = players[player]["status"]
    if status != "active":
        print(f"Cannot manage orders: {player} is {status}")

    dirty = False

    if player not in loaded_state.game.raw_orders:
        loaded_state.game.raw_orders[player] = []
    player_orders = loaded_state.game.raw_orders[player]

    while True:
        print(format_orders(player_orders, player))
        print("----------")
        print("1. Add an order")
        print("2. Delete an order")
        print("0. Go back")
        choice = input("Choose an option: ").strip()
        if choice == "1":
            new_order = input("Add a new order: ")
            player_orders.append(new_order)
            dirty = True
            continue

        elif choice == "2":
            idx_str = input("Choose the order number to delete").strip()
            order_idx = int(idx_str) - 1
            if 0 <= order_idx < len(player_orders):
                player_orders.pop(order_idx)
                dirty = True
                continue
            else:
                print("Not a valid order number")

        elif choice == "0":
            if dirty:
                save_player_orders(
                    loaded_state.game.game_meta["game_id"],
                    player,
                    player_orders,
                )
            return
        else:
            print("Invalid choice")
