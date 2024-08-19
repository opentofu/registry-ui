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
        <Paragraph className="text-balance">This technology preview contains documentation for a select few providers and namespaces in the OpenTofu registry.</Paragraph>
        <input
          autoFocus
          placeholder="Search providers, resources, or modules"
          className="relative block h-12 w-full max-w-[600px] appearance-none border border-transparent bg-gray-200 px-4 text-gray-800 placeholder:text-gray-700 focus:border-brand-700 focus:outline-none dark:bg-[#363A41] dark:text-white dark:placeholder:text-gray-300"
        />
      </main>
    </>
  );
}
