{ forCI ? false }:
let
  pkgs = import <nixpkgs> { };
in
with pkgs;
mkShell {
  buildInputs = [
    go
    nodejs
  ] ++ lib.lists.optional (!forCI) [
    gotools
    gopls
    vgo2nix
  ];
}
