{
  description = "Mines HSPC Website";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs { system = system; };
        in
        rec {
          packages.mineshspc = pkgs.buildGoModule rec {
            pname = "mineshspc.com";
            version = "unstable-2023-05-15";
            src = self;
            subPackages = [ "cmd/mineshspc" ];
            vendorSha256 = "sha256-x03AOeJj6zfcbt4n5UyoRrIIjF3rySoc/q4XOR0NLs0=";
          };
          packages.default = packages.mineshspc;

          devShells = {
            default = pkgs.mkShell {
              packages = with pkgs; [
                go_1_20
                gopls
                gotools
                pre-commit
              ];
            };
          };
        }
      ));
}
