# mdunwrap: Markdown Unwrapper

mdunwrap is a simple command-line tool for removing
[soft line breaks](https://spec.commonmark.org/0.30/#soft-line-breaks)
from [Markdown](https://en.wikipedia.org/wiki/Markdown).
This is helpful if you're copying Markdown from a text file
into a web form that renders soft line breaks as hard line breaks.

For example, it takes Markdown that looks like this:

```markdown
Hello,
World!
This is a single paragraph \
with some breaks.
```

into this:

```markdown
Hello, World! This is a single paragraph \
with some breaks.
```

## Getting Started

mdunwrap is canonically built using [Nix](https://nixos.org/).
After installing [Nix with Flakes support](https://zero-to-nix.com/start/install),
you can try out mdunwrap with:

```shell
nix run github:zombiezen/mdunwrap -- myfile.md
```

Or if you want to install it in your `PATH`:

```shell
nix profile install github:zombiezen/mdunwrap &&
mdunwrap myfile.md
```

## Caveats

- mdunwrap attempts to stay true to the original Markdown source,
  but may make modifications that don't change the rendered output.
- mdunwrap [does not support tables](https://github.com/zombiezen/mdunwrap/issues/1).

## License

[Apache 2.0](LICENSE)
