{ buildGoModule
, lib
}:

buildGoModule {
  pname = "mdunwrap";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-fgYQzbKAHY7okJxnas06c3uVeLmQFbkhHAk8QGnJWic=";

  meta = with lib; {
    description = "Markdown unwrapper";
    homepage = "https://github.com/zombiezen/mdunwrap";
    license = licenses.asl20;
    maintainers = with maintainers; [ zombiezen ];
  };
}
