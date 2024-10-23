{
  description = "Mines HSPC Website";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in rec {
        packages.mineshspc = pkgs.buildGo122Module {
          pname = "mineshspc.com";
          version = "unstable-2024-10-23";
          src = self;
          subPackages = [ "cmd/mineshspc" ];
          vendorHash = "sha256-0svb8gvvmvWh0F4fm4iqy9uMoqIBU+LSc36eRiDJ2cc=";
        };
        packages.default = packages.mineshspc;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gopls gotools pre-commit ];
        };
      }));
}
