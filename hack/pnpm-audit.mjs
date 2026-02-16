import { spawnSync } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import path from 'node:path'

const dir = process.argv[2]
if (!dir) {
  console.error('Usage: node hack/pnpm-audit.mjs <dir>')
  process.exit(2)
}

function readInstalledVersion(workdir, packageName) {
  const pkgPath = path.resolve(workdir, 'node_modules', ...packageName.split('/'), 'package.json')
  if (!existsSync(pkgPath)) return null
  try {
    const pkg = JSON.parse(readFileSync(pkgPath, 'utf8'))
    return typeof pkg.version === 'string' ? pkg.version : null
  } catch {
    return null
  }
}

function readInstalledDependencies(workdir, packageName) {
  const pkgPath = path.resolve(workdir, 'node_modules', ...packageName.split('/'), 'package.json')
  if (!existsSync(pkgPath)) return {}
  try {
    const pkg = JSON.parse(readFileSync(pkgPath, 'utf8'))
    return pkg?.dependencies && typeof pkg.dependencies === 'object' ? pkg.dependencies : {}
  } catch {
    return {}
  }
}

function readLatestVersion(packageName) {
  const out = spawnSync('pnpm', ['view', packageName, 'version'], { encoding: 'utf8' })
  const latest = (out.stdout ?? '').trim()
  if (out.status !== 0 || !latest) {
    const err = (out.stderr ?? '').trim()
    console.error(`[audit] Failed to resolve latest version of '${packageName}'`)
    if (err) console.error(err)
    process.exit(1)
  }
  return latest
}

function readLatestDependencies(packageName) {
  const out = spawnSync('pnpm', ['view', `${packageName}@latest`, 'dependencies', '--json'], { encoding: 'utf8' })
  const raw = (out.stdout ?? '').trim()
  if (out.status !== 0) {
    const err = (out.stderr ?? '').trim()
    console.error(`[audit] Failed to resolve latest dependencies of '${packageName}'`)
    if (err) console.error(err)
    process.exit(1)
  }
  if (!raw) return {}
  try {
    const deps = JSON.parse(raw)
    return deps && typeof deps === 'object' ? deps : {}
  } catch {
    console.error(`[audit] Failed to parse dependency json for '${packageName}'`)
    process.exit(1)
  }
}

function runAudit(workdir) {
  const audit = spawnSync('pnpm', ['-C', workdir, 'audit', '--json'], { encoding: 'utf8' })
  const stdout = (audit.stdout ?? '').trim()
  if (!stdout) {
    if (audit.stderr) process.stderr.write(audit.stderr)
    process.exit(audit.status ?? 1)
  }
  try {
    return JSON.parse(stdout)
  } catch {
    process.stdout.write(stdout + '\n')
    if (audit.stderr) process.stderr.write(audit.stderr)
    process.exit(audit.status ?? 1)
  }
}

function advisoryLabel(advisory) {
  const id = advisory.github_advisory_id ?? advisory.id ?? 'unknown'
  const sev = advisory.severity ?? 'unknown'
  const mod = advisory.module_name ?? 'unknown'
  const title = advisory.title ?? '(no title)'
  return `[${sev}] ${id} ${mod}: ${title}`
}

function advisoryLink(advisory) {
  const id = advisory?.github_advisory_id
  if (typeof id === 'string' && id.startsWith('GHSA-')) {
    return `https://github.com/advisories/${id}`
  }
  return advisory?.url ?? 'unknown'
}

function parseSemver(version) {
  const m = /^(\d+)\.(\d+)\.(\d+)/.exec(version.trim())
  if (!m) return null
  return [Number(m[1]), Number(m[2]), Number(m[3])]
}

function compareSemver(a, b) {
  for (let i = 0; i < 3; i++) {
    if (a[i] > b[i]) return 1
    if (a[i] < b[i]) return -1
  }
  return 0
}

function extractPatchedMin(patchedVersions) {
  if (typeof patchedVersions !== 'string') return null
  const m = /^\s*>=\s*([0-9]+\.[0-9]+\.[0-9]+)\s*$/.exec(patchedVersions)
  if (!m) return null
  return parseSemver(m[1])
}

function extractRangeMin(depRange) {
  if (typeof depRange !== 'string') return null
  const m = /([0-9]+\.[0-9]+\.[0-9]+)/.exec(depRange)
  if (!m) return null
  return parseSemver(m[1])
}

function advisoryPaths(advisory) {
  const findings = Array.isArray(advisory?.findings) ? advisory.findings : []
  return findings.flatMap((f) => (Array.isArray(f?.paths) ? f.paths : [])).filter((p) => typeof p === 'string')
}

function printAdvisorySummary(advisory) {
  const moduleName = advisory?.module_name ?? 'unknown'
  const patched = advisory?.patched_versions ?? 'unknown'
  const paths = [...new Set(advisoryPaths(advisory))]
  const pathsText = paths.length > 0 ? paths.join(', ') : 'unknown'

  console.log(`    📦 ${moduleName} patched ${patched} (${pathsText})`)
}

function printVersionComparison(item, advisory) {
  const moduleName = advisory?.module_name ?? 'unknown'
  const vulnerable = advisory?.vulnerable_versions ?? 'unknown'
  const currentParentVersion = item.installedParentVersion ?? 'unknown'
  const currentModuleRange = item.currentDepRange ?? 'unknown'
  const latestModuleRange = item.latestDepRange ?? 'unknown'

  console.log(`    - current: ${item.parentPackage}@${currentParentVersion} (${moduleName}@${currentModuleRange} ${vulnerable})`)
  console.log(`    - latest:  ${item.parentPackage}@${item.latestParentVersion} (${moduleName}@${latestModuleRange} ${vulnerable})`)
}

function parsePathSegments(pathValue) {
  return pathValue
    .split('>')
    .map((s) => s.trim())
    .filter((s) => s && s !== '.')
}

function extractParentPackages(advisory, moduleName) {
  const paths = advisoryPaths(advisory)
  const parents = new Set()
  for (const p of paths) {
    const seg = parsePathSegments(p)
    if (seg.length < 2) continue
    if (seg[seg.length - 1] !== moduleName) continue
    parents.add(seg[seg.length - 2])
  }
  return [...parents]
}

const latestCache = new Map()
function getLatestPackageInfo(packageName) {
  if (latestCache.has(packageName)) return latestCache.get(packageName)
  const latestVersion = readLatestVersion(packageName)
  const latestDeps = readLatestDependencies(packageName)
  const info = { latestVersion, latestDeps }
  latestCache.set(packageName, info)
  return info
}

const report = runAudit(dir)
const advisories = Object.values(report.advisories ?? {})

if (advisories.length === 0) {
  console.log(`[audit] ${dir}: 🟢 ok`)
  process.exit(0)
}

const ignored = []
const upgradeRequired = []
const blocking = []

for (const advisory of advisories) {
  const moduleName = advisory?.module_name
  if (typeof moduleName !== 'string') {
    blocking.push({ advisory, reason: 'module_name not found' })
    continue
  }

  const parentPackages = extractParentPackages(advisory, moduleName)
  if (parentPackages.length === 0) {
    blocking.push({ advisory, reason: 'cannot extract vulnerable dependency parent package from path' })
    continue
  }
  if (parentPackages.length > 1) {
    blocking.push({ advisory, reason: `multiple parent packages found (${parentPackages.join(', ')})` })
    continue
  }
  const parentPackage = parentPackages[0]

  const patchedMin = extractPatchedMin(advisory?.patched_versions)
  const { latestVersion: latestParentVersion, latestDeps: latestParentDeps } = getLatestPackageInfo(parentPackage)
  const latestDepRange = latestParentDeps[moduleName]
  const latestDepMin = extractRangeMin(latestDepRange)
  const installedParentVersion = readInstalledVersion(dir, parentPackage)
  const installedParentDeps = readInstalledDependencies(dir, parentPackage)
  const currentDepRange = installedParentDeps[moduleName]

  if (!patchedMin || !latestDepRange || !latestDepMin) {
    blocking.push({ advisory, reason: `cannot decide if latest ${parentPackage}@${latestParentVersion} resolves '${moduleName}'` })
    continue
  }

  if (compareSemver(latestDepMin, patchedMin) >= 0) {
    upgradeRequired.push({ advisory, parentPackage, installedParentVersion, latestParentVersion, latestDepRange, currentDepRange })
  } else {
    ignored.push({ advisory, parentPackage, installedParentVersion, latestParentVersion, latestDepRange, currentDepRange })
  }
}

if (blocking.length > 0) {
  console.log(`[audit] ${dir}: blocking advisories = ${blocking.length}`)
  for (const item of blocking) {
    console.log(`- ${advisoryLabel(item.advisory)}`)
    console.log(`  reason: ${item.reason}`)
  }
  process.exit(1)
}

if (upgradeRequired.length > 0) {
  console.error(`[audit] ${dir}: 🔴 resolvable vulnerabilities found (upgrade required)`)
  const upgradeTargets = new Map()
  for (const item of upgradeRequired) {
    upgradeTargets.set(item.parentPackage, item.latestParentVersion)
  }
  for (const [pkg, latest] of upgradeTargets.entries()) {
    const installed = readInstalledVersion(dir, pkg)
    if (installed) {
      console.error(`[audit] installed: ${pkg}@${installed}`)
    }
    console.error(`[audit] latest:    ${pkg}@${latest}`)
    console.error(`[audit] upgrade command: pnpm -C ${dir} up ${pkg}@${latest}`)
  }
  for (const item of upgradeRequired) {
    const sev = item.advisory?.severity ?? 'unknown'
    console.error(`- 🚨 [${sev}] resolvable ${advisoryLink(item.advisory)}`)
    printAdvisorySummary(item.advisory)
    printVersionComparison(item, item.advisory)
  }
  process.exit(1)
}

console.log(`[audit] ${dir}: 🟡 ignored (unresolvable vulnerabilities)`)
for (const item of ignored) {
  console.log(`  🔒 ${advisoryLink(item.advisory)}`)
  printAdvisorySummary(item.advisory)
  printVersionComparison(item, item.advisory)
}
process.exit(0)
