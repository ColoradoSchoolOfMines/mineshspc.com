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
          version = "unstable-2023-05-15";
          src = self;
          subPackages = [ "cmd/mineshspc" ];
          vendorHash = "sha256-6kIbPQgxqbTuwiE++jmGWNVM02WJP6iFoYAMncrmSgc=";
        };
        packages.default = packages.mineshspc;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go_1_20 gopls gotools pre-commit ];
        };
      }));
}
