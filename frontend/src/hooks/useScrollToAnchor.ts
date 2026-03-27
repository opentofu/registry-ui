import { useCallback } from "react";

export function useScrollToAnchor() {
  const scrollToAnchor = useCallback((id: string) => {
    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'center' });
      
      // Add flash animation after scroll
      element.style.transition = 'background-color 0.3s ease-in-out';
      element.style.backgroundColor = 'rgb(250 204 21 / 0.3)'; // Yellow with opacity
      
      // Remove the highlight after 2.5 seconds
      setTimeout(() => {
        element.style.backgroundColor = '';
        // Clean up the transition after animation
        setTimeout(() => {
          element.style.transition = '';
        }, 300);
      }, 2500);
    }
  }, []);

  return scrollToAnchor;
}