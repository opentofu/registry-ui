const dateFormat = new Intl.DateTimeFormat();

export function formatDate(date: string) {
  return dateFormat.format(new Date(date));
}
