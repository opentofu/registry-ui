import { PrismTheme } from "prism-react-renderer";

// TODO: finish the theme
export const theme: PrismTheme = {
  plain: {},
  styles: [
    {
      types: ["comment"],
      style: {
        color: "var(--syntax-comment)",
      },
    },
    {
      types: ["keyword"],
      style: {
        color: "var(--syntax-keyword)",
      },
    },
    {
      types: ["boolean"],
      style: {
        color: "var(--syntax-boolean)",
      },
    },
    {
      types: ["property"],
      style: {
        color: "var(--syntax-property)",
      },
    },
    {
      types: ["punctuation"],
      style: {
        color: "var(--syntax-punctuation)",
      },
    },
    {
      types: ["string"],
      style: {
        color: "var(--syntax-string)",
      },
    },
    {
      types: ["variable"],
      style: {
        color: "var(--syntax-variable)",
      },
    },
  ],
};
