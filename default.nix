# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "catbus-snapcast-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/catbus-snapcast";

  modSha256 = "166p21x59l1v0zh5j4cj1bcz9fppv1mvihag74vsij9v9x3w4i6l";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;

  meta = {
    homepage = "https://ethulhu.co.uk/catbus";
    licence = stdenv.lib.licenses.mit;
  };
}
