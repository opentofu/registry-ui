import { Paragraph } from "../Paragraph";

export function Footer() {
  return (
    <footer className="border-t border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-blue-950">
      <div className="mx-auto flex w-full max-w-4xl text-balance px-5 py-5 text-center text-sm">
        <Paragraph>
          Copyright © OpenTofu a Series of LF Projects, LLC and its
          contributors. Documentation materials are licensed under various open
          sources license from other authors, see the referenced license files
          for details. For web site terms of use, trademark policy, privacy
          policy and other project policies please see{" "}
          <a
            href="https://lfprojects.org/policies"
            target="_blank"
            rel="noreferrer noopener"
            className="underline"
          >
            lfprojects.org/policies
          </a>
          .
        </Paragraph>
      </div>
    </footer>
  );
}
