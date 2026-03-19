import { useCallback } from "react";
import { useNavigate } from "react-router";
import { definitions } from "@/api";
import { getDocumentationUrl } from "../utils/getDocumentationUrl";



interface UseSearchKeyboardProps {
  flatResults: definitions["SearchResultItem"][];
  selectedIndex: number;
  setSelectedIndex: (index: number) => void;
  setSelectedResult: (result: definitions["SearchResultItem"]) => void;
  resultsContainerRef: React.RefObject<HTMLDivElement>;
  setQuery: (query: string) => void;
  onClearSearch: () => void;
}

/**
 * Custom hook that handles keyboard navigation for search results.
 * 
 * Provides keyboard shortcuts for:
 * - Arrow Up/Down: Navigate through search results with visual feedback and auto-scrolling
 * - Enter: Navigate to the full documentation page for the selected result
 * - Escape: Close the search interface and clear the query
 * 
 * The hook manages result selection state and integrates with React Router for navigation.
 */
export function useSearchKeyboard({
  flatResults,
  selectedIndex,
  setSelectedIndex,
  setSelectedResult,
  resultsContainerRef,
  setQuery,
  onClearSearch,
}: UseSearchKeyboardProps) {
  const navigate = useNavigate();
  
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelectedIndex((prev) => {
          const newIndex = Math.min(prev + 1, flatResults.length - 1);
          setSelectedResult(flatResults[newIndex]);
          // Scroll to selected item
          const container = resultsContainerRef.current;
          const items = container?.querySelectorAll("[data-result-item]");
          if (items && items[newIndex]) {
            items[newIndex].scrollIntoView({
              block: "nearest",
              behavior: "smooth",
            });
          }
          return newIndex;
        });
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelectedIndex((prev) => {
          const newIndex = Math.max(prev - 1, 0);
          setSelectedResult(flatResults[newIndex]);
          // Scroll to selected item
          const container = resultsContainerRef.current;
          const items = container?.querySelectorAll("[data-result-item]");
          if (items && items[newIndex]) {
            items[newIndex].scrollIntoView({
              block: "nearest",
              behavior: "smooth",
            });
          }
          return newIndex;
        });
      } else if (e.key === "Escape") {
        onClearSearch();
      } else if (e.key === "Enter" && flatResults.length > 0) {
        e.preventDefault();
        const selectedResult = flatResults[selectedIndex];
        if (selectedResult) {
          const url = getDocumentationUrl(selectedResult);
          if (url) {
            navigate(url);
          }
        }
      }
    },
    [
      flatResults,
      selectedIndex,
      setSelectedIndex,
      setSelectedResult,
      resultsContainerRef,
      setQuery,
      onClearSearch,
      navigate,
    ],
  );

  return { handleKeyDown };
}