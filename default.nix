# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "catbus-snapcast-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/catbus-snapcast";

  modSha256 = "06ihmp5ykmr9gnq75m9l3lmgx1rhfsj2xasvdd1g3shm68r5975c";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;

  meta = {
    homepage = "https://ethulhu.co.uk/catbus";
    licence = stdenv.lib.licenses.mit;
  };
}
