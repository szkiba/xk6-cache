import { describe as _describe } from 'https://jslib.k6.io/expect/0.0.4/index.js'
import { expect } from 'https://cdnjs.cloudflare.com/ajax/libs/chai/4.3.7/chai.min.js'
import { search } from 'https://cdnjs.cloudflare.com/ajax/libs/jmespath/0.15.0/jmespath.min.js'
import qs from 'https://cdnjs.cloudflare.com/ajax/libs/qs/6.10.1/qs.min.js'

import { Rate } from 'k6/metrics'

export let errors = new Rate('errors')
export let options = {
  thresholds: { errors: [{ threshold: 'rate == 0.00', abortOnFail: true }] }
}

function describe (name, fn) {
  let success = _describe(name, fn)
  errors.add(!success)
}

export default function () {
  describe('expect', t => {
    expect('42').equal('42')
  })

  describe('jmespath', t => {
    let result = search(
      { foo: { bar: { baz: [0, 1, 2, 3, 4] } } },
      'foo.bar.baz[2]'
    )
    t.expect(result).as('result').toEqual(2)
  })

  describe('qs', t => {
    var obj = qs.parse('answer=42')
    t.expect(obj.answer).as('answer').toEqual('42')
  })
}
