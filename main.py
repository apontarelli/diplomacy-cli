from logic import load, save, validator, engine, state

def main():

    print("Starting game loop")
    load()
    validator()
    engine()
    state()
    save()

main()
