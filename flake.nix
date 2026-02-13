{
  description = "etu/llr";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      packages = flake-utils.lib.flattenTree {
        default = pkgs.buildGoModule (let
          version = "0.9.2.${nixpkgs.lib.substring 0 8 self.lastModifiedDate}.${self.shortRev or "dirty"}";
        in {
          pname = "llr";
          inherit version;

          src = ./.;

          vendorHash = "sha256-25a2WMXvabxUelm0OKVAN+YQ5WKb05Euq+ubWdvaXgQ=";
        });
      };

      devShells = flake-utils.lib.flattenTree {
        default = pkgs.mkShell {
          buildInputs = [
            pkgs.gnumake
            pkgs.delve # debugging
            pkgs.go # language
            pkgs.gopls # language server
          ];
        };
      };

      formatter = pkgs.alejandra;
    });
}
