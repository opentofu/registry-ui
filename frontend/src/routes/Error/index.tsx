import { useRouteError } from "react-router";
import { Header } from "@/components/Header";
import { Paragraph } from "@/components/Paragraph";
import PatternBg from "@/components/PatternBg";
import { is404Error } from "@/utils/errors";

export function Error() {
  const routeError = useRouteError() as Error;

  const is404 = is404Error(routeError);

  const title = is404 ? "Page Not Found" : "An Error Occurred";

  const message = is404
    ? "The page you are looking for does not exist."
    : "We're sorry, but an unexpected error occurred. Please try again later.";

  return (
    <>
      <PatternBg />
      <Header />
      <main className="container m-auto flex flex-col items-center gap-8 text-center pt-24">
        <h2 className="text-6xl font-bold">{title}</h2>
        <Paragraph className="text-balance">{message}</Paragraph>
        {import.meta.env.DEV && !!routeError.message && (
          <pre className="text-balance">{routeError.message}</pre>
        )}
      </main>
    </>
  );
}
