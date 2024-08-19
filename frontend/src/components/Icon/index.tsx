interface IconProps {
  path: string;
  className?: string;
  viewBox?: string;
  width?: number;
  height?: number;
}

export function Icon({ path, className, width = 24, height = 24 }: IconProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox={`0 0 ${width} ${height}`}
      aria-hidden="true"
      className={className}
    >
      <path fill="currentColor" d={path} />
    </svg>
  );
}
