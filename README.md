# xk6-cache

A k6 extension enables vendoring [Remote HTTP(s) modules](https://k6.io/docs/using-k6/modules#remote-http-s-modules) to single source control friendly file.

k6 supports importing [Remote HTTP(s) modules](https://k6.io/docs/using-k6/modules#remote-http-s-modules). The imported modules will be downloaded and executed at runtime. Relying on external servers is convenient for development but brittle in production. Production software should always vendor its dependencies.

Using xk6-cache this is done by pointing `$XK6_CACHE` to some project-local file at runtime, and similarly checking that into source control:

```bash
# Download the dependencies.
XK6_CACHE=vendor.k6c k6 run --out cache script.js

# Make sure the variable is set for any command which invokes the cache.
XK6_CACHE=vendor.k6c k6 run script.js

# Check the cache file into source control.
git add -u vendor.k6c
git commit
```

Built for [k6](https://go.k6.io/k6) using [xk6](https://github.com/grafana/xk6).

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [JavaScript API](#javascript-api)
- [Create cache file](#create-cache-file)
- [Use cache file](#use-cache-file)
- [How it works](#how-it-works)
- [Build](#build)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## JavaScript API

Usage of xk6-cache extension is transparent for k6 scripts. However sometimes good to see some metrics about cache usage (number of entries, cache hit count, cache miss count). 

You may call [measure()](docs/README.md#measure) function to enable cache metrics.

```JavaScript
import cache from "k6/x/cache";

export function setup() {
  cache.measure()
}
```

After cache metrics is enabled, the normal k6 metrics output will include the following three metrics:

```
     xk6_cache_entry_count...: 5
     xk6_cache_hit_count.....: 3
     xk6_cache_miss_count....: 0
```

## Create cache file

When cache file (pointed by `$XK6_CACHE`) is missing and you run k6 with `--out cache` flag then file will created.

```plain
$ XK6_CACHE=vendor.k6c k6 run --out cache script.js

          /\      |‾‾| /‾‾/   /‾‾/   
     /\  /  \     |  |/  /   /  /    
    /  \/    \    |     (   /   ‾‾\  
   /          \   |  |\  \ |  (‾)  | 
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: script.js
     output: cache (vendor.k6c)
```

If cache file already exists then `--out cache` has no effect, cache file remain untouch. This enable you to use `--out cache` flag always and simly delete cache file when you want to update it.

## Use cache file

Simly point `$XK6_CACHE` to existing cache file. Usage of `--out cache` flag has no effect if cache file exists (this is why you see dash instead of output file name).

```plain
$ XK6_CACHE=vendor.k6c k6 run --out cache script.js

          /\      |‾‾| /‾‾/   /‾‾/   
     /\  /  \     |  |/  /   /  /    
    /  \/    \    |     (   /   ‾‾\  
   /          \   |  |\  \ |  (‾)  | 
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: script.js
     output: cache (-)
```

## How it works

Well, it's a bit tricky. Since k6 extension API has no lifecycle hooks and the [k6 module loader](https://github.com/k6io/k6/tree/master/loader) is not usable from extensions, xk6-cache hijacks `http.DefaultTransport` and do the cache checking and cache recording as a [http.RoundTripper](https://golang.org/pkg/net/http/#RoundTripper) interceptor.

It is not a risk for k6 scripts, because `k6/http` module doesn't rely on `http.DefaultTransport`. It would be nice to have an API for extensions to implement custom module loaders.

The cache is a single [FlatBuffers](https://google.github.io/flatbuffers/) file which is store URLs and the downloaded modules only (sorted by URL). This mean the file is almost a text file and source control friendly.

## Build

To build a `k6` binary with this extension, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

Then:

1. Install `xk6`:
  ```bash
  $ go install go.k6.io/xk6/cmd/xk6@latest
  ```

2. Build the binary:
  ```bash
  $ xk6 build --with github.com/szkiba/xk6-cache@latest
  ```
