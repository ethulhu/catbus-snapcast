# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "catbus-snapcast-${version}";
  version = "latest";
  goPackagePath = "go.eth.moe/catbus-snapcast";

  modSha256 = "0s31r2ccf5n53l0qpg8z94rbjb1aakkkkzhc6hh32ggmnp8blmsl";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;

  meta = {
    homepage = "https://ethulhu.co.uk/catbus";
    licence = stdenv.lib.licenses.mit;
  };
}
