import type { Timestamp } from '@bufbuild/protobuf/wkt'

/**
 * Formats a duration in seconds to m:ss, m:ss.f, or m:ss.ff
 * @param seconds Duration in seconds
 * @param precision Number of decimal places (0, 1, or 2)
 */
export function formatDuration(seconds: number, precision: number = 0): string {
  if (seconds < 0) return ''
  
  const minutes = Math.floor(seconds / 60)
  const wholeSeconds = Math.floor(seconds % 60)
  
  let timeStr = `${minutes}:${wholeSeconds.toString().padStart(2, '0')}`
  
  if (precision > 0) {
    // Use Math.floor to consistently truncate, avoiding rounding inconsistencies
    // e.g. 65.99 should be 1:05.9, not 1:06.0, matching the 1:05 base
    const factor = Math.pow(10, precision)
    // Avoid (seconds % 1) as it introduces precision errors (e.g. 1.2 % 1 ~= 0.1999)
    // Instead, shift the decimal point first, then modulo by factor to get the fractional digits.
    const fractionalPartVal = Math.floor((seconds * factor) % factor)
    const fractionalPart = fractionalPartVal.toString().padStart(precision, '0')
    timeStr += `.${fractionalPart}`
  }
  
  return timeStr
}

/**
 * Calculates the duration in seconds between two timestamps
 */
export function calculateDurationSeconds(start: Timestamp, end: Timestamp): number {
  const startSeconds = Number(start.seconds) + Number(start.nanos) / 1e9
  const endSeconds = Number(end.seconds) + Number(end.nanos) / 1e9
  return Math.max(0, endSeconds - startSeconds)
}

/**
 * Generates a map of formatted time strings for a list of items.
 * Automatically increases precision (up to 2 decimal places) for items that would
 * otherwise have identical formatted strings.
 * 
 * @param items List of objects with an id and duration in seconds
 * @returns Map of item id to formatted string
 */
export function getFormattedTimes(items: { id: string; seconds: number }[]): Map<string, string> {
  const results = new Map<string, string>()
  
  // Helper to group items by their formatted string at a given precision
  const groupItems = (currentItems: { id: string; seconds: number }[], precision: number) => {
    const groups = new Map<string, { id: string; seconds: number }[]>()
    
    for (const item of currentItems) {
      const str = formatDuration(item.seconds, precision)
      if (!groups.has(str)) {
        groups.set(str, [])
      }
      groups.get(str)!.push(item)
    }
    return groups
  }

  // Process groups recursively
  const processGroups = (currentItems: { id: string; seconds: number }[], precision: number) => {
    // If we've reached max precision (2), just output whatever we have
    if (precision > 2) {
      for (const item of currentItems) {
        results.set(item.id, formatDuration(item.seconds, 2))
      }
      return
    }

    const groups = groupItems(currentItems, precision)
    
    for (const [formattedStr, group] of groups) {
      // Check if this formatted string represents a "zero" time at current precision
      const isZeroValue = 
        (precision === 0 && formattedStr === '0:00') ||
        (precision === 1 && formattedStr === '0:00.0')

      if (group.length === 1 && precision < 2 && !isZeroValue) {
        // Unique at this precision and not a zero-value we want to expand
        results.set(group[0]!.id, formattedStr)
      } else if (precision === 2) {
         // At max precision, use this string for all items in the group
         for (const item of group) {
           results.set(item.id, formattedStr)
         }
      } else {
        // Collision exists OR it's a zero-value that needs more precision
        processGroups(group, precision + 1)
      }
    }
  }

  processGroups(items, 0)
  return results
}
