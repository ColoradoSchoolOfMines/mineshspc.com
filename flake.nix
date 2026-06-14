{
  description = "Mines HSPC Website";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      rec {
        packages.mineshspc = pkgs.buildGoModule {
          pname = "mineshspc.com";
          version = "unstable-2026-01-11";
          src = self;
          subPackages = [ "cmd/mineshspc" ];
          vendorHash = "sha256-yEXplnaY4rYZ0WUTy3S92PGizXPJ9j6wOHly70R1lMY=";
        };
        packages.default = packages.mineshspc;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            gcc
            go
            pre-commit
          ];
        };
      }
    ));
}
