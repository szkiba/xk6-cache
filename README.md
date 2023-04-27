# xk6-cache

A k6 extension enables vendoring [Remote HTTP(s) modules](https://k6.io/docs/using-k6/modules#remote-http-s-modules) to single source control friendly file.

k6 supports importing [Remote HTTP(s) modules](https://k6.io/docs/using-k6/modules#remote-http-s-modules). The imported modules will be downloaded and executed at runtime. Relying on external servers is convenient for development but brittle in production. Production software should always vendor its dependencies.

Using xk6-cache this is done by pointing `$XK6_CACHE` to some project-local file at runtime, and similarly checking that into source control:

```bash
# Download the dependencies.
XK6_CACHE=vendor.eml k6 run --out cache script.js

# Make sure the variable is set for any command which invokes the cache.
XK6_CACHE=vendor.eml k6 run script.js

# Check the cache file into source control.
git add -u vendor.eml
git commit
```

## Create cache file

When cache file (pointed by `$XK6_CACHE`) is missing and you run k6 with `--out cache` flag then file will created.

```plain
$ XK6_CACHE=vendor.eml k6 run --out cache script.js

          /\      |‾‾| /‾‾/   /‾‾/   
     /\  /  \     |  |/  /   /  /    
    /  \/    \    |     (   /   ‾‾\  
   /          \   |  |\  \ |  (‾)  | 
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: script.js
     output: cache (vendor.eml)
```

If cache file already exists then `--out cache` has no effect, cache file remain untouch. This enable you to use `--out cache` flag always and simly delete cache file when you want to update it.

## Use cache file

Simly point `$XK6_CACHE` to existing cache file. Usage of `--out cache` flag has no effect if cache file exists (this is why you see dash instead of output file name).

```plain
$ XK6_CACHE=vendor.eml k6 run --out cache script.js

          /\      |‾‾| /‾‾/   /‾‾/   
     /\  /  \     |  |/  /   /  /    
    /  \/    \    |     (   /   ‾‾\  
   /          \   |  |\  \ |  (‾)  | 
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: script.js
     output: cache (vendor.eml)
```

## How it works

Well, it's a bit tricky. Since k6 extension API has no lifecycle hooks and the [k6 module loader](https://github.com/k6io/k6/tree/master/loader) is not usable from extensions, xk6-cache hijacks `http.DefaultTransport` and do the cache checking and cache recording as a [http.RoundTripper](https://golang.org/pkg/net/http/#RoundTripper) interceptor.

It is not a risk for k6 scripts, because `k6/http` module doesn't rely on `http.DefaultTransport`. It would be nice to have an API for extensions to implement custom module loaders.

The cache is a single plain text file which is store URLs and the downloaded modules only (sorted by URL). This mean the file is  a text file and source control friendly. The file format is standard email text format, so if you choose `.eml` as file extension, you can view the content with an email client (like Mozilla Thinderbird).

<details><summary>Example</summary>
<p>

For the following script file ...

```js
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export default function () {
  const randomUUID = uuidv4();
  console.log(randomUUID);
}
```

... will be generated this cache file:

```eml
Content-Type: multipart/mixed; boundary=______________________________o_o______________________________
Subject: xk6-cache

--______________________________o_o______________________________
Content-Type: text/plain; charset=utf-8

This is xk6-cache's standard email format cache file that can be viewed with an email client such as Mozilla Thunderbird. Modules stored as email attachments.
--______________________________o_o______________________________
Content-Disposition: attachment; filename="https://jslib.k6.io/k6-utils/1.4.0/index.js"
Content-Length: 4974
Content-Location: https://jslib.k6.io/k6-utils/1.4.0/index.js?_k6=1
Content-Type: text/javascript

(()=>{"use strict";var t={n:r=>{var e=r&&r.__esModule?()=>r.default:()=>r;return .......__esModule&&Object.defineProperty(w,"__esModule",{value:!0})})();
//# sourceMappingURL=index.js.map
--______________________________o_o______________________________--

```

> **Note**
> The long, minified JavaScript code replaced with `.......` in the example above.

</p>
</details>

## Download

You can download pre-built k6 binaries from [Releases](https://github.com/szkiba/xk6-cache/releases/) page. Check [Packages](https://github.com/szkiba/xk6-cache/pkgs/container/xk6-cache) page for pre-built k6 Docker images.

## Build

You can build the k6 binary on various platforms, each with its requirements. The following shows how to build k6 binary with this extension on GNU/Linux distributions.

### Prerequisites

You must have the latest Go version installed to build the k6 binary. The latest version should match [k6](https://github.com/grafana/k6#build-from-source) and [xk6](https://github.com/grafana/xk6#requirements).

- [Git](https://git-scm.com/) for cloning the project
- [xk6](https://github.com/grafana/xk6) for building k6 binary with extensions

### Install and build the latest tagged version

1. Install `xk6`:

   ```shell
   go install go.k6.io/xk6/cmd/xk6@latest
   ```

2. Build the binary:

   ```shell
   xk6 build --with github.com/szkiba/xk6-cache@latest
   ```

> **Note**
> You can always use the latest version of k6 to build the extension, but the earliest version of k6 that supports extensions via xk6 is v0.43.0. The xk6 is constantly evolving, so some APIs may not be backward compatible.

### Build for development

If you want to add a feature or make a fix, clone the project and build it using the following commands. The xk6 will force the build to use the local clone instead of fetching the latest version from the repository. This process enables you to update the code and test it locally.

```bash
git clone git@github.com:szkiba/xk6-cache.git && cd xk6-cache
xk6 build --with github.com/szkiba/xk6-cache@latest=.
```

## Docker

You can also use pre-built k6 image within a Docker container. In order to do that, you will need to execute something like the following:

**Linux**

```plain
docker run -v $(pwd):/scripts -e XK6_CACHE=vendor.eml -it --rm ghcr.io/szkiba/xk6-cache:latest run --out=cache /scripts/script.js
```

**Windows**

```plain
docker run -v %cd%:/scripts -e XK6_CACHE=vendor.eml -it --rm ghcr.io/szkiba/xk6-cache:latest run --out=cache /scripts/script.js
```
