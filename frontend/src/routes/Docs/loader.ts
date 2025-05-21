import { LoaderFunction } from "react-router-dom";
import { NotFoundPageError } from "@/utils/errors";
import * as prodUtils from "./utils" with { type: "macro" };
import * as devUtils from "./utils";

const slugPathMap = import.meta.env.PROD
  ? prodUtils.getSlugPathMap()
  : devUtils.getSlugPathMap();

export const docsLoader: LoaderFunction = async ({ params }) => {
  let slug = "";

  if (params.section) {
    slug += params.section;

    if (params.subsection) {
      slug += "/" + params.subsection;
    }
  }

  const normalizedSlug = slug.replace(/[^a-zA-Z0-9/-]/g, "");

  const document = slugPathMap[normalizedSlug];

  if (!document) {
    throw new NotFoundPageError();
  }

  return document;
};
