import { Search } from "@/components/Search";
import { Header } from "../../components/Header";
import { Paragraph } from "../../components/Paragraph";
import PatternBg from "../../components/PatternBg";

export function Home() {
  return (
    <>
      <PatternBg />
      <Header />
      <main className="container m-auto flex flex-col items-center gap-8 text-center">
        <h2 className="text-6xl font-bold">
          Documentation for OpenTofu Providers and Modules
        </h2>
        <Paragraph className="text-balance">
          This technology preview contains documentation for a select few
          providers and namespaces in the OpenTofu registry.
        </Paragraph>
        <Search
          size="large"
          placeholder="Search providers, resources, or modules"
        />
      </main>
    </>
  );
}
