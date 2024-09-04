import { Search } from "@/components/Search";
import { Header } from "../../components/Header";
import { Paragraph } from "../../components/Paragraph";
import PatternBg from "../../components/PatternBg";
import { Footer } from "@/components/Footer";

export function Home() {
  return (
    <>
      <PatternBg />
      <Header />
      <main className="container m-auto flex flex-col items-center text-center">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 340 368"
          className="-mt-24 size-24"
        >
          <path
            fill="#0D1A2B"
            fillRule="evenodd"
            d="M182.26 3.88a25.74 25.74 0 0 0-24.8 0L30.4 73.73l-.31.17-16.74 9.2A25.74 25.74 0 0 0 0 105.66v157.1c0 9.39 5.12 18.03 13.34 22.56l128.28 70.5.5.27 15.34 8.43a25.74 25.74 0 0 0 24.8 0l15.39-8.45.45-.25 128.28-70.5a25.74 25.74 0 0 0 13.34-22.56v-157.1c0-9.4-5.11-18.04-13.34-22.56l-16.67-9.16-.37-.21L182.26 3.88Zm8.17 180.32 118.9-65.35.37-.2 1.42-.79a5.94 5.94 0 0 1 8.8 5.2v122.29a5.94 5.94 0 0 1-8.8 5.2l-1.1-.6-.68-.39-118.91-65.36ZM30.08 118.68l.31.17L149.3 184.2 30.4 249.56l-.67.38-1.12.61a5.94 5.94 0 0 1-8.8-5.2V123.07a5.94 5.94 0 0 1 8.8-5.2l1.47.8Zm269.9-27.5L188.56 29.95a5.94 5.94 0 0 0-8.8 5.2v132.31l120.21-66.07a5.94 5.94 0 0 0 0-10.2Zm-260.2 10.23 120.18 66.05V34.98a5.94 5.94 0 0 0-8.8-5.03L39.78 91.17a5.94 5.94 0 0 0 0 10.24Zm.15 175.92c-4-2.2-4.1-7.85-.31-10.23l120.34-66.15v132.49a5.94 5.94 0 0 1-8.56 5.16L39.93 277.33Zm139.83 55.93v-132.3l120.36 66.15a5.94 5.94 0 0 1-.32 10.22l-111.46 61.26a5.94 5.94 0 0 1-8.58-5.15v-.18Z"
            clipRule="evenodd"
          />
          <path
            fill="#E7C200"
            d="M167 21.23a5.94 5.94 0 0 1 5.72 0L299.8 91.07a5.94 5.94 0 0 1 0 10.42l-127.08 69.84a5.94 5.94 0 0 1-5.72 0L39.93 101.5a5.94 5.94 0 0 1 0-10.42L167 21.23Z"
          />
          <path
            fill="#FFDA18"
            d="M19.8 123.06a5.94 5.94 0 0 1 8.8-5.2l128.28 70.5a5.94 5.94 0 0 1 3.08 5.2v139.7a5.94 5.94 0 0 1-8.8 5.2l-128.28-70.5a5.94 5.94 0 0 1-3.08-5.2v-139.7Z"
          />
          <path
            fill="#fff"
            d="M311.12 117.86a5.94 5.94 0 0 1 8.8 5.2v139.7a5.94 5.94 0 0 1-3.08 5.2l-128.28 70.5a5.94 5.94 0 0 1-8.8-5.2v-139.7a5.94 5.94 0 0 1 3.08-5.2l128.28-70.5Z"
          />
          <path
            fill="#0D1A2B"
            d="m73.68 232.66-.02.2-30.06-15.82v-.2c.7-9.35 8-13.4 16.3-9.03 8.3 4.37 14.48 15.5 13.78 24.85ZM121 259.98l-.02.2-30.07-15.82.02-.2c.7-9.35 7.99-13.4 16.3-9.03 8.3 4.37 14.46 15.5 13.77 24.85Z"
          />
        </svg>
        <h2 className="mt-5 max-w-4xl text-balance text-6xl font-bold leading-tight">
          Documentation for OpenTofu Providers and Modules
        </h2>
        <Paragraph className="mb-7 mt-5 text-balance">
          This technology preview contains documentation for a select few
          providers, namespaces, and modules in the OpenTofu registry.
        </Paragraph>
        <Search
          size="large"
          placeholder="Search providers, resources, or modules"
        />
      </main>
      <Footer />
    </>
  );
}
