{
  description = "Mines HSPC Website";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ = {
      url = "github:a-h/templ/v0.2.697";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, templ }:
    (flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays =
            [ (final: prev: { inherit (templ.packages.${system}) templ; }) ];
        };
      in rec {
        packages.mineshspc = pkgs.buildGo122Module {
          pname = "mineshspc.com";
          version = "unstable-2024-05-21";
          src = self;
          subPackages = [ "cmd/mineshspc" ];
          vendorHash = "sha256-qlDZkMlvzrRhE/GdEz4hIQgk4BI2YrvOLazAFqs8xuY=";

          preBuild = ''
            ${pkgs.templ}/bin/templ generate
          '';
        };
        packages.default = packages.mineshspc;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gopls gotools pre-commit pkgs.templ ];
        };
      }));
}
