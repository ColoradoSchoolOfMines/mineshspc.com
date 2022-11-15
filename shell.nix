{ forCI ? false }:
let
  pkgs = import <nixpkgs> { };
in
with pkgs;
mkShell {
  buildInputs = [
    go_1_19
  ] ++ lib.lists.optional (!forCI) [
    gotools
    gopls
  ];
}
