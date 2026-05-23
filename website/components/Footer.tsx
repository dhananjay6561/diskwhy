import Link from "next/link";

export default function Footer() {
  return (
    <footer className="border-t border-line py-9">
      <div className="max-w-[1160px] mx-auto px-6 flex flex-wrap items-center justify-between gap-4">
        <div className="font-mono text-[0.875rem] text-ink3">
          <span className="text-brand">disk</span>why
        </div>

        <ul className="flex gap-6 list-none">
          {[
            { label: "home", href: "/" },
            { label: "showcase", href: "/showcase" },
            { label: "github", href: "https://github.com/dhananjay6561/diskwhy", external: true },
            { label: "readme", href: "https://github.com/dhananjay6561/diskwhy/blob/main/README.md", external: true },
            { label: "schema", href: "https://github.com/dhananjay6561/diskwhy/blob/main/SCHEMA_CHANGELOG.md", external: true },
          ].map(({ label, href, external }) =>
            external ? (
              <li key={label}>
                <a
                  href={href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-[0.8rem] text-ink3 hover:text-ink2 transition-colors"
                >
                  {label}
                </a>
              </li>
            ) : (
              <li key={label}>
                <Link href={href} className="text-[0.8rem] text-ink3 hover:text-ink2 transition-colors">
                  {label}
                </Link>
              </li>
            )
          )}
        </ul>

        <div className="text-[0.8rem] text-ink3">MIT License</div>
      </div>
    </footer>
  );
}
