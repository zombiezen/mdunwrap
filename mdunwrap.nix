{ deno
, cacert
, stdenv
, lib
}:

let
  entrypoint = "mdunwrap.ts";
  testFiles = [
    "mdunwrap_test.ts"
  ];
  vendorHash = "sha256-1XxJtHqjtqWUaLoXA5lQWA/dnXIIPXnca+KT15aXm3Q=";

  vendor = stdenv.mkDerivation {
    name = "mdunwrap-vendor";
    outputHash = vendorHash;
    outputHashMode = "recursive";

    src = ./.;

    nativeBuildInputs = [ deno cacert ];

    impureEnvVars = lib.fetchers.proxyImpureEnvVars ++ [
      "SOCKS_SERVER"
    ];

    dontConfigure = true;
    doCheck = false;
    dontFixup = true;

    buildPhase = ''
      runHook preBuild
      export DENO_DIR="$TMPDIR/deno"
      deno vendor --output=vendor ${entrypoint} ${builtins.concatStringsSep " " testFiles}
      mkdir -p vendor
      runHook postBuild
    '';

    installPhase = ''
      runHook preInstall
      cp -r --reflink=auto vendor $out
      runHook postInstall
    '';
  };
in

stdenv.mkDerivation {
  pname = "mdunwrap";
  version = "0.1.0";

  src = ./.;

  nativeBuildInputs = [ deno ];

  dontConfigure = true;
  dontFixup = true;

  buildPhase = ''
    runHook preBuild
    export DENO_DIR="$TMPDIR/deno"
    deno compile \
      --output=mdunwrap \
      --no-remote \
      --import-map=${vendor}/import_map.json \
      --allow-read --allow-write \
      ${entrypoint}
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    mkdir -p $out/bin
    cp --reflink=auto mdunwrap $out/bin/
    runHook postInstall
  '';

  doCheck = true;
  checkPhase = ''
    runHook preCheck
    deno test \
      --no-remote \
      --import-map=${vendor}/import_map.json \
      --allow-read=testdata \
      ${builtins.concatStringsSep " " testFiles}
    runHook postCheck
  '';

  passthru.vendor = vendor;

  meta = with lib; {
    description = "Markdown unwrapper";
    homepage = "https://github.com/zombiezen/mdunwrap";
    license = licenses.asl20;
    maintainers = with maintainers; [ zombiezen ];
  };
}
