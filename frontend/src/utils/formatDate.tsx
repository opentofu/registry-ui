const dateFormat = new Intl.DateTimeFormat();

export function formatDate(date: string) {
  return dateFormat.format(new Date(date));
}

export function formatDateTag(date: string) {
  return <time dateTime={date}>{formatDate(date)}</time>;
}
