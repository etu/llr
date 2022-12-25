{ pkgs ? import <nixpkgs> {}, ... }:

pkgs.buildGoModule {
  pname = "llr";
  version = "20221225";

  src = ./.;

  vendorSha256 = "14x9ikn9jfldvs8dznzf3xabqw0d9c20gzqggmiy4pca0s9bc34z";
}
