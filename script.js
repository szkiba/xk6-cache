/**
 * MIT License
 *
 * Copyright (c) 2021 Iván Szkiba
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

import { describe } from "https://jslib.k6.io/expect/0.0.4/index.js";
import { search } from "https://cdnjs.cloudflare.com/ajax/libs/jmespath/0.15.0/jmespath.min.js";

import cache from "k6/x/cache";

export function setup() {
  cache.measure();
}

export default function () {
  describe("jmespath", (t) => {
    const result = search({ foo: { bar: { baz: [0, 1, 42, 3, 4] } } }, "foo.bar.baz[2]");
    t.expect(result).as("answer").toEqual(42);
  });
}
