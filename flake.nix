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
        {
          packages.mineshspc = pkgs.buildGoModule rec {
            pname = "mineshspc.com";
            version = "unstable-2023-05-15";

            src = ./.;

            vendorSha256 = "sha256-fBEEXT0nvAagXpL4Prmrua8g6iNnLyyt1DJcphVbtuM=";
          };
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
