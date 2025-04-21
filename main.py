from logic import start_game, load, save, validator, engine, state

def main():

    print("Starting game loop")
    start_game()
    validator()
    engine()

main()
