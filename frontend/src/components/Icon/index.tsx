interface IconProps {
  path: string;
  className?: string;
  viewBox?: string;
  width?: number;
  height?: number;
  title?: string;
}

export function Icon({
  title,
  path,
  className,
  width = 24,
  height = 24,
}: IconProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox={`0 0 ${width} ${height}`}
      aria-hidden={title ? undefined : true}
      className={className}
    >
      {title ? <title>{title}</title> : null}
      <path fill="currentColor" d={path} />
    </svg>
  );
}
