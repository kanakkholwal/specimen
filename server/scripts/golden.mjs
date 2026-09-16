import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'

const chunks = process.argv.slice(2, -2)
const resultPath = process.argv.at(-2)
const outDir = process.argv.at(-1)

if (!chunks.length || !resultPath || !outDir) {
  console.error('usage: node golden.mjs <chunk.js...> <result.json> <outDir>')
  process.exit(2)
}

const factories = []
const sandbox = {
  console,
  globalThis: null,
  document: undefined,
  self: {},
  window: undefined,
  TURBOPACK: {
    push(entry) {
      // Entry layout is not a strict id/factory alternation, so take every function.
      for (const item of entry) if (typeof item === 'function') factories.push(item)
    },
  },
}
sandbox.globalThis = sandbox
sandbox.self = sandbox

const ctx = vm.createContext(sandbox)
for (const c of chunks) {
  vm.runInContext(fs.readFileSync(c, 'utf8'), ctx, { filename: path.basename(c) })
}

// Stubbed module context. Symbol traps must stay primitive-friendly or module init
// throws before the generator exports are registered.
const stub = new Proxy(function () {}, {
  get(_t, k) {
    if (k === Symbol.toPrimitive) return () => ''
    if (k === Symbol.iterator) return function* () {}
    if (k === Symbol.toStringTag) return 'Stub'
    if (k === 'toString' || k === 'valueOf') return () => ''
    if (typeof k === 'symbol') return undefined
    return stub
  },
  apply: () => stub,
  construct: () => stub,
  has: () => true,
})

const exported = {}
function moduleCtx() {
  return {
    s(spec) {
      if (Array.isArray(spec[0])) for (const one of spec) exported[one[0]] = one[2]
      else exported[spec[0]] = spec[2]
    },
    i: () => stub,
    r: () => stub,
    v: () => stub,
    n: () => stub,
    m: {},
    e: {},
  }
}

let loaded = 0
for (const factory of factories) {
  const src = String(factory)
  if (!/generateDesignMd|generateCssVariables|generateTailwindConfig|generateDesignTokens/.test(src)) continue
  try {
    factory(moduleCtx())
    loaded++
  } catch (err) {
    console.error(`module failed: ${err.message}`)
  }
}

const names = Object.keys(exported).filter((k) => typeof exported[k] === 'function')
console.error(`loaded ${loaded} module(s); generators: ${names.join(', ')}`)

const result = JSON.parse(fs.readFileSync(resultPath, 'utf8'))
fs.mkdirSync(outDir, { recursive: true })

// Names follow the site's own FORMAT_META table; every generator that accepts a
// variant is emitted in both, since the export bar offers Compact and Extended.
const outputs = [
  ['DESIGN.compact.md', () => exported.generateDesignMd(result, 'compact')],
  ['DESIGN.extended.md', () => exported.generateDesignMd(result, 'extended')],
  ['variables.compact.css', () => exported.generateCssVariables(result, 'compact')],
  ['variables.extended.css', () => exported.generateCssVariables(result, 'extended')],
  ['theme.compact.css', () => exported.generateTailwindConfig(result, 'compact')],
  ['theme.extended.css', () => exported.generateTailwindConfig(result, 'extended')],
  ['tokens.compact.json', () => exported.generateTokensJson(result, 'compact')],
  ['tokens.extended.json', () => exported.generateTokensJson(result, 'extended')],
  ['design-system.json', () => exported.generateDesignJson(result)],
  ['prompt.txt', () => exported.generatePromptSnapshotText(result)],
]

let wrote = 0
for (const [name, run] of outputs) {
  try {
    const text = run()
    if (typeof text !== 'string') throw new Error(`returned ${typeof text}`)
    fs.writeFileSync(path.join(outDir, name), text)
    console.error(`${name}: ${text.length} chars`)
    wrote++
  } catch (err) {
    console.error(`${name}: FAILED ${err.message}`)
  }
}
process.exit(wrote === outputs.length ? 0 : 1)
