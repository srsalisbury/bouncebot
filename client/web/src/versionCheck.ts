/**
 * Checks if the running app version matches the latest deployed version.
 * If they differ, forces a one-time page reload to pick up the new assets.
 */
export async function checkVersion(): Promise<void> {
  if (__APP_VERSION__ === 'dev') return

  // Guard against infinite reload loops: if we already reloaded for this
  // baked-in version and the mismatch persists (e.g. cached index.html),
  // don't reload again.
  const key = 'version-check-reloaded'
  if (sessionStorage.getItem(key) === __APP_VERSION__) return

  try {
    const basePath = window.APP_CONFIG?.BASE_PATH || '/'
    const res = await fetch(`${basePath}version.json?t=${Date.now()}`)
    if (!res.ok) return

    const { version } = await res.json()
    if (version && version !== __APP_VERSION__) {
      console.log(`Version mismatch: running ${__APP_VERSION__}, server has ${version}. Reloading.`)
      sessionStorage.setItem(key, __APP_VERSION__)
      window.location.reload()
    }
  } catch {
    // Network error or offline, skip check
  }
}
