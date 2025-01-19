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
        packages.mineshspc = pkgs.buildGoModule {
          pname = "mineshspc.com";
          version = "unstable-2024-12-19";
          src = self;
          subPackages = [ "cmd/mineshspc" ];
          vendorHash = "sha256-aMutiq/y90lx5lIVV4foQnPp2dKuvnD5OHS0rPi5/sg=";
        };
        packages.default = packages.mineshspc;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gotools pre-commit ];
        };
      }));
}
