import styles from "./styles.module.css";

export default function PatternBg() {
  return (
    <div
      className="absolute bottom-0 left-0 right-0 -z-10 h-[600px]"
      aria-hidden
    >
      <div className="absolute inset-0 z-10 bg-gradient-to-b from-gray-50 to-transparent dark:from-blue-950" />
      <div className={`${styles.bg} h-full w-full opacity-5`} />
    </div>
  );
}
