# quickjs

![github ci](https://github.com/newdash/quickjs/workflows/github%20ci/badge.svg)

Go bindings to [QuickJS](https://bellard.org/quickjs/): a fast, small, and embeddable [ES2020](https://tc39.github.io/ecma262/) JavaScript interpreter.

These bindings are a WIP and do not match full parity with QuickJS' API, though expose just enough features to be usable. The version of QuickJS that these bindings bind to may be located [here](version.h).

## Usage

```bash
$ go get github.com/newdash/quickjs
```

## Guidelines

1. Free `quickjs.Runtime` and `quickjs.Context` once you are done using them.
2. Free `quickjs.Value`'s returned by `Eval()` and `EvalFile()`. All other values do not need to be freed, as they get garbage-collected.
3. You may access the stacktrace of an error returned by `Eval()` or `EvalFile()` by casting it to a `*quickjs.Error`.
4. Make new copies of arguments should you want to return them in functions you created.
5. Make sure to call `runtime.LockOSThread()` to ensure that QuickJS always operates in the exact same thread.

## Build

`linux/MacOS` build is tested with CI, and for the `Windows`, the `libdl` library is required.

## License

`QuickJS` is released under the MIT license.

`QuickJS bindings` are copyright `Kenta Iwasaki`, with code copyright `Fabrice Bellard` and `Charlie Gordon`.

