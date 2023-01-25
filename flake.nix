{
  description = "etu/llr";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, flake-utils, nixpkgs, ... }: flake-utils.lib.eachDefaultSystem (system: let
    pkgs = import nixpkgs { inherit system; };
  in {
    packages = flake-utils.lib.flattenTree {
      default = pkgs.buildGoModule (let
        version = "1.0.${nixpkgs.lib.substring 0 8 self.lastModifiedDate}.${self.shortRev or "dirty"}";
      in {
        pname = "llr";
        inherit version;

        src = ./.;

        vendorSha256 = "14x9ikn9jfldvs8dznzf3xabqw0d9c20gzqggmiy4pca0s9bc34z";
      });
    };

    devShells = flake-utils.lib.flattenTree {
      default = { pkgs, ... }: pkgs.mkShell {
        buildInputs = [
          pkgs.gnumake # For the Makefile
          pkgs.go      # For building the project
        ];
      };
    };
  });
}
