import { useRouteError } from "react-router-dom";
import { Header } from "../../components/Header";
import { Paragraph } from "../../components/Paragraph";
import PatternBg from "../../components/PatternBg";
import { NotFoundPageError } from "@/utils/errors";

export function Error() {
  const routeError = useRouteError() as Error;

  const title =
    routeError instanceof NotFoundPageError
      ? "Page Not Found"
      : "An Error Occurred";

  const message =
    routeError instanceof NotFoundPageError
      ? "The page you are looking for does not exist."
      : "We're sorry, but an unexpected error occurred. Please try again later.";

  return (
    <>
      <PatternBg />
      <Header />
      <main className="container m-auto flex flex-col items-center gap-8 text-center">
        <h2 className="text-6xl font-bold">{title}</h2>
        <Paragraph className="text-balance">{message}</Paragraph>
        {!!routeError.message && (
          <pre className="text-balance">{routeError.message}</pre>
        )}
      </main>
    </>
  );
}
