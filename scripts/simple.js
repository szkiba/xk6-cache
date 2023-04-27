import { describe } from "https://jslib.k6.io/expect/0.0.4/index.js";

export default function () {
  describe("expect", (t) => {
    const answer = 42
    t.expect(answer).as("answer").toEqual(42)
  })
}
