/**
 * Translates technical API error messages into user-friendly messages.
 */
export function translateJoinRoomError(error: unknown): string {
  const message = error instanceof Error ? error.message : 'Failed to join room'
  const lower = message.toLowerCase()

  if (lower.includes('not found') || lower.includes('invalid')) {
    return 'Room not found. Please check the Room ID.'
  }
  if (lower.includes('single player')) {
    return 'This is a solo game and cannot be joined.'
  }

  return message
}
