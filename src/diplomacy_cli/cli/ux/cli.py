from diplomacy_cli.core.logic.state import load_state, start_game
from diplomacy_cli.core.logic.storage import delete_game, list_games

from .pretty import format_state


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
        games = list_games()
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
        print("2. Delete game")
        print("0. Back to menu")
        print("--------------------")
        choice = input("Choose an option: ").strip()
        if choice == "1":
            state = load_state(game_id)
            print(format_state(state))
        elif choice == "2":
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
