/**
 * Checks if the running app version matches the latest deployed version.
 * If they differ, forces a page reload to pick up the new assets.
 */
export async function checkVersion(): Promise<void> {
  if (__APP_VERSION__ === 'dev') return

  try {
    const res = await fetch(`/version.json?t=${Date.now()}`)
    if (!res.ok) return

    const { version } = await res.json()
    if (version && version !== __APP_VERSION__) {
      console.log(`Version mismatch: running ${__APP_VERSION__}, server has ${version}. Reloading.`)
      window.location.reload()
    }
  } catch {
    // Network error or offline, skip check
  }
}
