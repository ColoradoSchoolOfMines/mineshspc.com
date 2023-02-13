{ forCI ? false }:
let
  pkgs = import <nixpkgs> { };
in
with pkgs;
mkShell {
  buildInputs = [
    go_1_19
    pre-commit
  ] ++ lib.lists.optional (!forCI) [
    gotools
    gopls
  ];
}
