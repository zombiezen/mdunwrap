{ buildGoModule
, lib
}:

buildGoModule {
  pname = "mdunwrap";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-g0KKT6vNd4csP01gYTKgEtQxSF2MWccvkObWp7LCT9w=";

  meta = with lib; {
    description = "Markdown unwrapper";
    homepage = "https://github.com/zombiezen/mdunwrap";
    license = licenses.asl20;
    maintainers = with maintainers; [ zombiezen ];
  };
}
