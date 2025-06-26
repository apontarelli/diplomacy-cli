{
  description = "A development environment for the Diplomacy CLI project";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
        python = pkgs.python312;
        python-with-packages = python.withPackages (ps: with ps; [
          uv
          ruff
        ]);
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.pyright
            python-with-packages
          ];
          shellHook = ''
            # Define the virtual environment directory for uv and other tools
            export VIRTUAL_ENV="$(pwd)/.venv"

            # If the venv doesn't exist, create it and install dependencies.
            if [ ! -f "$VIRTUAL_ENV/pyvenv.cfg" ]; then
              echo "Setting up Python environment..."
              uv venv
              uv pip install -e ".[dev]"
            fi

            # Add the virtual environment's bin to the PATH
            export PATH="$VIRTUAL_ENV/bin:$PATH"
          '';
        };
      }
    );
}
