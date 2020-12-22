// Format a Date object into a human-readable string.
// This string strives to be readable and unambiguous across different cultures,
// which is we use a ISO8601-like format.
export function formatDateTime(date: Date): string {
  // https://stackoverflow.com/a/58633686
  return date.toLocaleTimeString( 'sv-SE', {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
  });
}

// Parses a date time string and formats into a human-readable string.
// This string strives to be readable and unambiguous across different cultures,
// which is we use a ISO8601-like format.
export function formatDateTimeString(dateStr: string): string {
  return formatDateTime(new Date(dateStr));
}

export function capitalizeFirstLetter(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
