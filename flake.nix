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
          version = "unstable-2026-09-20";
          src = self;
          subPackages = [ "cmd/mineshspc" ];
          vendorHash = "sha256-M63AfE8MgRoZZxOkFnxp2/CffoYBANzL/qq/XURnBZw=";
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
