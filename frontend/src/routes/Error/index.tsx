import { useRouteError } from "react-router-dom";
import { Header } from "../../components/Header";
import { Paragraph } from "../../components/Paragraph";
import PatternBg from "../../components/PatternBg";

export function Error() {
  const routeError = useRouteError() as Error;
  return (
    <>
      <PatternBg />
      <Header />
      <main className="container m-auto flex flex-col items-center gap-8 text-center">
        <h2 className="text-6xl font-bold">An Error Occurred</h2>
        <Paragraph className="text-balance">
          We're sorry, but an unexpected error occurred. Please try again later.
        </Paragraph>
        <pre className="text-balance">{routeError.message}</pre>
      </main>
    </>
  );
}
