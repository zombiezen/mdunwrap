{ buildGoModule
, lib
}:

buildGoModule {
  pname = "mdunwrap";
  version = "0.1.0";

  src = ./.;

  # vendorHash = lib.fakeHash;
  vendorHash = null;

  meta = with lib; {
    description = "Markdown unwrapper";
    homepage = "https://github.com/zombiezen/mdunwrap";
    license = licenses.asl20;
    maintainers = with maintainers; [ zombiezen ];
  };
}
