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
          pyright
        ]);
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            python-with-packages
          ];
          shellHook = ''
            echo "Entering Diplomacy CLI development environment..."
            export VIRTUAL_ENV="$(pwd)/.venv"
            export PATH="$VIRTUAL_ENV/bin:$PATH"
          '';
        };
      }
    );
}
