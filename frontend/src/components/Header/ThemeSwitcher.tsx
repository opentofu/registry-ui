import { useState } from "react";
import { sun } from "../../icons/sun";
import { Icon } from "../Icon";
import { moon } from "../../icons/moon";

export function ThemeSwitcher() {
  const [isDark, setIsDark] = useState(
    document.documentElement.classList.contains("dark"),
  );

  const toggleTheme = () => {
    const newTheme = isDark ? "light" : "dark";
    localStorage.setItem("theme", newTheme);

    const newValue = !isDark;
    setIsDark(newValue);
    document.documentElement.classList.toggle("dark", newValue);
  };

  return (
    <button
      aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      onClick={toggleTheme}
    >
      <Icon path={isDark ? sun : moon} className="size-6" />
    </button>
  );
}
