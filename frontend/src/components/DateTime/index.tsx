const dateFormat = new Intl.DateTimeFormat();

interface DateTimeProps {
  value: string;
}

export function DateTime({ value }: DateTimeProps) {
  const dateTimeObj = dateFormat.format(new Date(value));
  return <time dateTime={value}>{dateTimeObj}</time>;
}
