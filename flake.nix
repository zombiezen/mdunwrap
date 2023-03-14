{
  description = "Markdown unwrapper";

  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = "flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
      };
    in
    {
      # packages.default = pkgs.callPackage ./mdunwrap.nix {};

      # apps.default = {
      #   type = "app";
      #   program = "${self.packages.${system}.default}/bin/mdunwrap";
      # };

      devShells.default = pkgs.mkShell {
        packages = [
          pkgs.deno
        ];
      };
    });
}
