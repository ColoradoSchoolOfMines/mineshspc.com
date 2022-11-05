{ forCI ? false }:
let
  pkgs = import <nixpkgs> { };
in
with pkgs;
mkShell {
  buildInputs = [
    go_1_19
    hugo
  ] ++ lib.lists.optional (!forCI) [
    gotools
    gopls
  ];
}
