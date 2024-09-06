import { LoaderFunction } from "react-router-dom";
import { NotFoundPageError } from "@/utils/errors";
import { slugPathMap } from "./utils" with { type: "macro" };

export const docsLoader: LoaderFunction = async ({ params }) => {
  const { "*": slug = "" } = params;
  const normalizedSlug = slug.replace(/[^a-zA-Z0-9/-]/g, "");

  const document = slugPathMap[normalizedSlug];

  if (!document) {
    throw new NotFoundPageError();
  }

  return document;
};
