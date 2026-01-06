import { ConnectError, Code } from '@connectrpc/connect'

/**
 * Error types that indicate the server actively rejected the request
 * (room/player doesn't exist), vs network errors where we should retry.
 */
export type ServerRejectionReason = 'room_not_found' | 'player_not_found'

/**
 * Check if an error is a server rejection (room/player deleted) vs a network error.
 * Returns the rejection reason if it's a server rejection, null otherwise.
 *
 * Network errors should NOT trigger state cleanup - the user might just have
 * a bad connection and the room might still exist.
 */
export function getServerRejectionReason(error: unknown): ServerRejectionReason | null {
  if (!(error instanceof ConnectError)) {
    // Not a Connect error - likely a network error
    return null
  }

  if (error.code === Code.NotFound) {
    return 'room_not_found'
  }

  if (error.code === Code.PermissionDenied || error.code === Code.Unauthenticated) {
    return 'player_not_found'
  }

  // Other Connect errors (Unavailable, Internal, etc.) are not rejections
  return null
}

/**
 * Check if an error indicates the room was garbage collected or doesn't exist.
 */
export function isRoomNotFoundError(error: unknown): boolean {
  return getServerRejectionReason(error) === 'room_not_found'
}

/**
 * Check if an error indicates the player is not in the room.
 */
export function isPlayerNotFoundError(error: unknown): boolean {
  return getServerRejectionReason(error) === 'player_not_found'
}

/**
 * Check if we should clear stored state based on the error.
 * Only returns true for active server rejections, not network errors.
 */
export function shouldClearStateOnError(error: unknown): boolean {
  return getServerRejectionReason(error) !== null
}
