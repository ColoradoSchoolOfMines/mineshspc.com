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
          version = "unstable-2024-04-11";
          src = self;
          subPackages = [ "cmd/mineshspc" ];
          vendorHash = "sha256-Tum0t4e9T0vP7Y5EE69+2MgR+QsVJ5QzQvM+QldlS8o=";
        };
        packages.default = packages.mineshspc;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go_1_22 gopls gotools pre-commit ];
        };
      }));
}
