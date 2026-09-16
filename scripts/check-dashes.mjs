import fs from 'node:fs'
import path from 'node:path'

const ROOTS = ['src', 'scripts', '.notes']
const EXTS = new Set(['.ts', '.js', '.mjs', '.svelte', '.css', '.html', '.md', '.json'])
const DASHES = /[\u2014\u2013]/g

function walk(dir, out = []) {
  if (!fs.existsSync(dir)) return out
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name)
    if (e.isDirectory()) {
      if (e.name === 'node_modules' || e.name === '.svelte-kit') continue
      walk(p, out)
    } else if (EXTS.has(path.extname(e.name))) out.push(p)
  }
  return out
}

const hits = []
for (const file of ROOTS.flatMap((r) => walk(r))) {
  const lines = fs.readFileSync(file, 'utf8').split(/\r?\n/)
  lines.forEach((line, i) => {
    if (DASHES.test(line)) hits.push(`${file}:${i + 1}: ${line.trim().slice(0, 100)}`)
    DASHES.lastIndex = 0
  })
}

if (hits.length) {
  for (const h of hits) console.log(h)
  console.log(`\n${hits.length} em/en dash(es) found`)
  process.exit(1)
}
console.log('dash scan: clean')
