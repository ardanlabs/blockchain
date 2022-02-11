{
  description = "A proof-of-concept blockchain written for pedagogical purposes.";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system}.pkgs;
        appConstructor = pkg: binary: flake-utils.lib.mkApp { drv = pkg; name = binary; };
      in
      rec {
        packages = {
          ardan-blockchain = pkgs.buildGoModule {
            pname = "ardan-blockchain";
            version = "v0.0.1"; # no tags...

            src = ./.;

            vendorSha256 = null; # vendor/ folder exists

            meta = with pkgs.lib; {
              description = "A proof-of-concept blockchain written for pedagogical purposes.";
              homepage = https://github.com/ardanlabs/blockchain;
              license = licenses.apache;
              maintainers = with maintainers; [ johnrichardrinehart ];
              platforms = platforms.linux ++ platforms.darwin;
            };
          };
        };

        defaultPackage = packages.ardan-blockchain;

        apps = nixpkgs.lib.genAttrs [ "logfmt" "node" "wallet" ] (name: appConstructor packages.ardan-blockchain name);

        defaultApp = apps.node;
      }
    );
}
