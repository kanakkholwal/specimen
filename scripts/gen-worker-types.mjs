import { execFileSync } from 'node:child_process'
import fs from 'node:fs'

const out = 'src/worker-configuration.d.ts'
execFileSync('bunx', ['wrangler', 'types', '--path', out], { stdio: 'inherit', shell: true })

// wrangler points mainModule at the built worker, which would pull build output into the
// type program. The binding types are all we need.
const src = fs.readFileSync(out, 'utf8')
const stripped = src.replace(/^\s*mainModule:\s*typeof import\([^)]*\);\s*$/m, '')
fs.writeFileSync(out, stripped)
console.log(stripped === src ? 'mainModule not present' : 'stripped mainModule reference')
