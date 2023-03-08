{ buildGoModule
, lib
}:

buildGoModule {
  pname = "mdunwrap";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-Q4K/5HZq15182SAU0pJIjH0enC2o8ycxdxmY/WyXjtc=";

  meta = with lib; {
    description = "Markdown unwrapper";
    homepage = "https://github.com/zombiezen/mdunwrap";
    license = licenses.asl20;
    maintainers = with maintainers; [ zombiezen ];
  };
}
