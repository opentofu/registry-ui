import { HTTPError } from "ky";

export class NotFoundPageError extends Error {}

export function is404Error(error: unknown) {
  return (
    error instanceof NotFoundPageError ||
    (error instanceof HTTPError && error.response.status === 404)
  );
}
