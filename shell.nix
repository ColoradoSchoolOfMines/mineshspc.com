{ forCI ? false }:
let
  pkgs = import <nixpkgs> { };
in
with pkgs;
mkShell {
  buildInputs = [
    go
    hugo
  ] ++ lib.lists.optional (!forCI) [
    gotools
    gopls
  ];
}
