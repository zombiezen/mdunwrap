{
  description = "Markdown unwrapper";

  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = "flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
  let
    supportedSystems = [
      flake-utils.lib.system.x86_64-linux
      flake-utils.lib.system.aarch64-linux
      flake-utils.lib.system.x86_64-darwin
      flake-utils.lib.system.aarch64-darwin
    ];
  in
    flake-utils.lib.eachSystem supportedSystems (system:
    let
      pkgs = import nixpkgs { inherit system; };
    in
    {
      packages.default = pkgs.callPackage ./mdunwrap.nix {
        buildGoModule = pkgs.buildGo120Module;
      };

      apps.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/mdunwrap";
      };

      devShells.default = pkgs.mkShell {
        packages = [
          pkgs.go_1_20
        ];
      };
    });
}
