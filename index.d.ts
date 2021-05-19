/**
 * xk6-cache enables vendoring [Remote HTTP(s) modules](https://k6.io/docs/using-k6/modules#remote-http-s-modules) to single source control friendly file.
 *
 * k6 supports importing [Remote HTTP(s) modules](https://k6.io/docs/using-k6/modules#remote-http-s-modules). The imported modules will be downloaded and executed at runtime. Relying on external servers is convenient for development but brittle in production.
 * Production software should always vendor its dependencies.
 *
 * Using xk6-cache this is done by pointing `$XK6_CACHE` to some project-local file at runtime, and similarly checking that into source control:
 *
 * ```bash
 * # Download the dependencies.
 * XK6_CACHE=vendor.k6c k6 run --out cache script.js
 *
 * # Make sure the variable is set for any command which invokes the cache.
 * XK6_CACHE=vendor.k6c k6 run script.js
 *
 * # Check the cache file into source control.
 * git add -u vendor.k6c
 * git commit
 * ```
 *
 * Usage of xk6-cache extension is transparent for k6 scripts. However sometimes good to see some metrics about cache usage (number of entries, cache hit count, cache miss count).
 *
 * You may call measure() function to enable cache metrics.
 *
 * ```JavaScript
 * import cache from "k6/x/cache";
 *
 * export function setup() {
 * cache.measure()
 * }
 * ```
 *
 * After cache metrics is enabled, the normal k6 metrics output will include the following three metrics:
 *
 * ```
 *     xk6_cache_entry_count...: 5
 *     xk6_cache_hit_count.....: 3
 *     xk6_cache_miss_count....: 0
 * ```
 *
 */

/**
 * Enable cache related metrics.
 *
 * @param prefix use as a metrics name prefix instead of default `xk6_cache`
 * @returns true if success, false otherwise
 */
export function measure(prefix?: string): boolean;
