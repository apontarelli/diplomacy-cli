from logic import start_game, load_state, save, validator, engine, state

def main():

    print("Starting game loop")
    print(load_state("new_game"))
    validator()
    engine()

main()
