import os
from logic.state import start_game, load_state
from logic.storage import load, list_saved_games, delete_game

def main_menu():
    while True:
        print("\n--- Diplomacy CLI ---")
        print("1. Start new game")
        print("2. View saved games")
        print("3. Exit")

        choice = input("Choose an option: ").strip()
        if choice == "1":
            start_new_game()
        elif choice == "2":
            view_saved_games()
        elif choice == "3":
            print("Exiting...")
            break
        else:
            print("Invalid choice.")

def start_new_game():
    game_id = input("Enter a game name: ").strip()
    variant = "classic"
    if game_id:
        start_game(variant, game_id)
    else:
        print("Game name cannot be blank")

def view_saved_games():
    while True:
        games = list_saved_games()
        if not games:
            print("No saved games")
            return

        for i, game in enumerate(games):
            print(f"{i + 1}. {game}")
        print(f"0. Back to menu")

        try:
            selection = int(input("Select a save: "))
            if selection == 0:
                return
            selected_game = games[selection - 1]
            manage_save(selected_game)
            return
        except (ValueError, IndexError):
            print("Invalid selection.")
    

def manage_save(game_id):
    while True:
        print(f"--- {game_id} ---")
        print("1. Print state")
        print("2. Delete game")
        print("0. Back to menu")

        choice = input("Choose an option: ").strip()
        if choice == "1":
            state = load_state(game_id)
            print(state)
        elif choice == "2":
            confirm = input(f"Type exact game_id {game_id} to confirm delete: ").strip()
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
