{ buildGoModule
, lib
}:

buildGoModule {
  pname = "mdunwrap";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-VObcEfmPvTZ2RfUa+/w9F1P3MpDTDrNqrLyG60lROf4=";

  meta = with lib; {
    description = "Markdown unwrapper";
    homepage = "https://github.com/zombiezen/mdunwrap";
    license = licenses.asl20;
    maintainers = with maintainers; [ zombiezen ];
  };
}
