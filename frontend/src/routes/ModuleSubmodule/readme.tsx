import readme from "./readme.md?raw";
import { Markdown } from "../../components/Markdown";

export function ModuleSubmoduleReadme() {
  return <Markdown text={readme} />;
}
