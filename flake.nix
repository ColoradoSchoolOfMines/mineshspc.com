{
  description = "Mines HSPC Website";
  inputs = {
    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs-unstable, flake-utils }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs-unstable { system = system; };
        in
        {
          packages.mineshspc = pkgs.buildGoModule rec {
            pname = "mineshspc.com";
            version = "unstable-2023-05-15";

            src = ./.;

            vendorSha256 = "sha256-4/KzlfllQzHLiTJupPSOj0v9jCYqhcRMoxCm9fiRgdg=";
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
