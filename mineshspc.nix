{ lib, fetchFromGitHub, buildGoModule, olm }:

buildGoModule rec {
  pname = "mineshspc";
  version = "unstable-2023-05-15";

  src = ./.;

  vendorSha256 = "sha256-4/KzlfllQzHLiTJupPSOj0v9jCYqhcRMoxCm9fiRgdg=";
}
