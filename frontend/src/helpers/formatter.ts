export function formatDate(date: string | Date) {
  return dateTimeFormatter(date);
}

export function formatDateTime(date: string | Date | null) {
  if (date === null) {
    return "";
  }

  return dateTimeFormatter(date, {
    hour: "numeric",
    minute: "numeric",
    second: "numeric",
    hour12: false,
  });
}

export function timeAgo(utcTimestamp: number): string {
  const seconds = Math.floor(Date.now() / 1000 - utcTimestamp);
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} minute${minutes !== 1 ? "s" : ""} ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} hour${hours !== 1 ? "s" : ""} ago`;
  const days = Math.floor(hours / 24);
  return `${days} day${days !== 1 ? "s" : ""} ago`;
}

function dateTimeFormatter(date: string | Date, customOptions?: Intl.DateTimeFormatOptions) {
  const defaultOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  };

  const options = { ...defaultOptions, ...customOptions };

  const formatter = new Intl.DateTimeFormat("de-DE", options);

  const dateObj = typeof date === "string" ? new Date(date) : date;

  return formatter.format(dateObj);
}
