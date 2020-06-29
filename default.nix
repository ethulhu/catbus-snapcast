# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "catbus-snapcast-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/catbus-snapcast";

  modSha256 = "09jafsc871raczzd5gjjybrcc93w7fnsm5w0spqjjnkhhpgbiw2a";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;

  meta = {
    homepage = "https://ethulhu.co.uk/catbus";
    licence = stdenv.lib.licenses.mit;
  };
}
