import fs from 'node:fs'
import path from 'node:path'

const MAX_LINES = 2
const ROOTS = ['src', 'scripts']
const EXTS = new Set(['.ts', '.js', '.mjs', '.svelte'])

const args = new Set(process.argv.slice(2))
const all = args.has('--all')

function walk(dir, out = []) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      if (entry.name === 'node_modules' || entry.name === 'testdata' || entry.name === 'data') continue
      walk(p, out)
    } else if (EXTS.has(path.extname(entry.name))) {
      out.push(p)
    }
  }
  return out
}

const files = ROOTS.filter((r) => fs.existsSync(r)).flatMap((r) => walk(r))
const problems = []

// A directive is machine-readable, not prose, so the length rule does not apply.
const DIRECTIVE = /^\s*\/\/\s*(go:|eslint-|@ts-|nolint|lint:|Code generated )/
const BANNER = /^\s*\/\/\s*[-=*#~_]{6,}\s*$/

for (const file of files) {
  const lines = fs.readFileSync(file, 'utf8').split(/\r?\n/)
  let start = -1
  let count = 0

  const flush = (end) => {
    if (count > MAX_LINES) {
      problems.push(`${file}:${start + 1}: comment block is ${count} lines, max ${MAX_LINES}`)
    }
    // A comment block that opens the file and is followed by a blank line is a header block.
    if (start === 0 && count > 0 && !/^\s*(\/\/\s*Package |\/\/\s*Command |import|package)/.test(lines[end] ?? '')) {
      if ((lines[end] ?? '').trim() === '') {
        problems.push(`${file}:1: file-header comment block`)
      }
    }
    start = -1
    count = 0
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const isComment = /^\s*\/\//.test(line)
    if (isComment && DIRECTIVE.test(line)) {
      if (count) flush(i)
      continue
    }
    if (isComment && BANNER.test(line)) {
      // A single named section divider is allowed; a bare rule is not.
      if (!/[A-Za-z]/.test(line)) problems.push(`${file}:${i + 1}: ASCII banner`)
      if (count) flush(i)
      continue
    }
    if (isComment) {
      if (count === 0) start = i
      if (line.replace(/^\s*\/\/\s*/, '').trim() === '' && count > 0) {
        problems.push(`${file}:${i + 1}: blank comment line`)
      }
      count++
      continue
    }
    if (count) flush(i)
  }
  if (count) flush(lines.length)
}

if (problems.length) {
  for (const p of problems) console.log(p)
  console.log(`\n${problems.length} problem(s) in ${files.length} file(s)`)
  process.exit(1)
}
console.log(`comment gate: clean (${files.length} files, scope ${all ? 'all' : 'default'})`)
