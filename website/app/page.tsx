import Link from "next/link";
import Nav from "@/components/Nav";
import Footer from "@/components/Footer";
import Terminal from "@/components/Terminal";
import CopyButton from "@/components/CopyButton";

const INSTALL_CMD = "go install github.com/dhananjay6561/diskwhy@latest";

const HERO_TERMINAL = `<span class="tp">~</span> <span class="tc">diskwhy scan</span>

<span class="tsi">  Scanning /Users/dj (quick) ...</span>

<span class="th">  FOUND 4.2 GB across 8 items</span>
<span class="ts">  ────────────────────────────────────────────────</span>
<span class="tsi">  1  </span><span class="tcat">node_modules  </span><span class="tsi">~/old-app/node_modules     </span><span class="tsz">1.8 GB  </span><span class="tu">unused</span>
<span class="tsi">  2  </span><span class="tcat">docker        </span><span class="tsi">3 images · 2 volumes       </span><span class="tsz">923 MB  </span><span class="tz">—</span>
<span class="tsi">  3  </span><span class="tcat">xcode_derived </span><span class="tsi">~/Library/Developer/Xcode  </span><span class="tsz">631 MB  </span><span class="tst">stale</span>
<span class="tsi">  4  </span><span class="tcat">node_modules  </span><span class="tsi">~/blog/node_modules        </span><span class="tsz">412 MB  </span><span class="tr">recent</span>
<span class="tsi">  5  </span><span class="tcat">pycache       </span><span class="tsi">~/scripts/__pycache__      </span><span class="tsz">248 MB  </span><span class="tst">stale</span>
<span class="tsi">  6  </span><span class="tcat">brew_cache    </span><span class="tsi">~/Library/Caches/Homebrew  </span><span class="tsz">178 MB  </span><span class="tu">unused</span>
<span class="tsi">  7  </span><span class="tcat">git_objects   </span><span class="tsi">~/monorepo/.git            </span><span class="tsz"> 94 MB  </span><span class="tst">stale</span>
<span class="tsi">  8  </span><span class="tcat">logs          </span><span class="tsi">~/Library/Logs             </span><span class="tsz"> 81 MB  </span><span class="tst">stale</span>
<span class="ts">  ────────────────────────────────────────────────</span>

<span class="tsi">  Run: </span><span class="tc">diskwhy clean --all --dry-run</span>

<span class="tp">~</span> <span class="cursor">▋</span>`;

const GH_ICON = (
  <svg viewBox="0 0 16 16" fill="currentColor" className="w-3.5 h-3.5 shrink-0">
    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
  </svg>
);

const CATEGORIES = [
  { name: "node_modules", desc: "npm, yarn, pnpm dependency trees" },
  { name: "git_objects", desc: "Loose objects and pack files in .git — runs git gc, not raw delete" },
  { name: "docker", desc: "Dangling images and unused volumes via Docker API" },
  { name: "pycache", desc: "Python __pycache__ directories and .pyc bytecode files" },
  { name: "pip_cache", desc: "pip download and wheel cache (~/.cache/pip)" },
  { name: "npm_cache", desc: "npm global cache in ~/.npm" },
  { name: "brew_cache", desc: "Homebrew downloads cache (macOS)" },
  { name: "xcode_derived", desc: "Xcode DerivedData and build artifacts (macOS)" },
  { name: "apt_cache", desc: "apt package cache in /var/cache/apt (Linux)" },
  { name: "snap_cache", desc: "Snap package cache (Linux)" },
  { name: "logs", desc: "Log files matching *.log and *.log.* patterns" },
  { name: "trash", desc: "Items in system Trash / ~/.Trash" },
];

const STALENESS = [
  { color: "#22c55e", name: "active",  desc: "modified < 7 days ago — never deleted" },
  { color: "#6ee7b7", name: "recent",  desc: "accessed < 30 days ago" },
  { color: "#f59e0b", name: "stale",   desc: "not accessed in 30–90 days" },
  { color: "#f87171", name: "unused",  desc: "not accessed in > 90 days" },
];

const FEATURES = [
  {
    icon: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-4 h-4"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>,
    title: "Active items are never deleted",
    body: <>Items modified in the last 7 days are marked <C>active</C> and unconditionally skipped by the clean phase. Cache categories (pycache, brew, npm, pip, etc.) are always safe regardless of staleness — they regenerate on demand.</>,
  },
  {
    icon: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-4 h-4"><circle cx="12" cy="12" r="10"/><path d="M12 8v4l3 3"/></svg>,
    title: "Staleness frozen at scan time",
    body: <>Access time, mtime, and sentinel files (<C>package.json</C>, <C>go.mod</C>, <C>Cargo.toml</C>) are evaluated once during the scan. The clean command never re-reads them — what you saw in the preview is exactly what gets acted on.</>,
  },
  {
    icon: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-4 h-4"><path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/></svg>,
    title: "Docker API, not exec.Command",
    body: <>Docker integration uses the <C>github.com/docker/docker</C> SDK directly — no <C>{`exec.Command("docker", "images")`}</C>. Queries containers, images, and volumes from the daemon socket. Silently skips if Docker is unavailable.</>,
  },
  {
    icon: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-4 h-4"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>,
    title: "JSON output, schema versioned",
    body: <>Every command supports <C>--json</C> with a stable <C>schema_version: 1</C> contract. Pipe to jq, wire into dashboards, or integrate with your own tooling. See the <a href="https://github.com/dhananjay6561/diskwhy/blob/main/SCHEMA_CHANGELOG.md" target="_blank" rel="noopener noreferrer" className="text-brand underline decoration-brand/30">schema changelog</a>.</>,
  },
  {
    icon: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-4 h-4"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>,
    title: "Git repos run gc, not delete",
    body: <>The <C>git_objects</C> category runs <C>git gc --prune=now</C> instead of nuking .git. Loose objects are pruned and pack files are consolidated. The repo stays intact. The disk space comes back.</>,
  },
  {
    icon: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-4 h-4"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>,
    title: "Interactive shell built in",
    body: <>Run <C>diskwhy</C> with no arguments to open a REPL. Type <C>/scan</C>, <C>/clean --all --dry-run</C>, <C>/help</C>. Adaptive to dark and light terminals. No readline dependency.</>,
  },
];

function C({ children }: { children: React.ReactNode }) {
  return <code className="bg-panel3 border border-line rounded px-1 py-px text-[.85em] text-brand">{children}</code>;
}

function Label({ children }: { children: string }) {
  return <span className="inline-block font-mono text-[0.7rem] text-brand tracking-widest uppercase mb-2.5">{children}</span>;
}

function InstallStrip() {
  return (
    <div className="flex items-center bg-panel border border-line rounded-md overflow-hidden">
      <code className="flex-1 px-4 py-3 font-mono text-[0.8125rem] text-ink2 whitespace-nowrap overflow-hidden text-ellipsis">
        <span className="text-ink3">$</span> {INSTALL_CMD}
      </code>
      <CopyButton text={INSTALL_CMD} />
    </div>
  );
}

export default function Home() {
  return (
    <>
      <Nav active="home" />

      {/* HERO */}
      <section className="pt-[148px] pb-24">
        <div className="max-w-[1160px] mx-auto px-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-14 items-center">
            <div>
              <div className="inline-flex items-center gap-2 px-3 py-1 bg-brand/10 border border-brand/20 rounded-full font-mono text-[0.75rem] text-brand mb-6">
                <span className="w-1.5 h-1.5 bg-brand rounded-full shrink-0" />
                go CLI &nbsp;·&nbsp; macOS &amp; Linux &nbsp;·&nbsp; MIT
              </div>

              <h1 className="text-[clamp(2.25rem,4.5vw,3.75rem)] font-semibold tracking-[-0.045em] leading-[1.08] text-white mb-5">
                Your disk is full.<br />
                <span className="text-brand">But why?</span>
              </h1>

              <p className="text-[1.0625rem] text-ink2 leading-relaxed mb-8 max-w-[420px]">
                diskwhy scans your drive and finds what&apos;s actually using space — not just &ldquo;big file on Desktop&rdquo;, but node_modules, Docker images, Xcode caches, stale Python bytecode. Then removes them safely.
              </p>

              <div className="mb-5 max-w-[500px]">
                <InstallStrip />
              </div>

              <div className="flex items-center gap-2.5 flex-wrap">
                <a
                  href="https://github.com/dhananjay6561/diskwhy"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 px-4 py-2.5 bg-brand border border-brand rounded-md text-[0.875rem] font-medium text-black hover:bg-green-600 hover:border-green-600 transition-colors"
                >
                  {GH_ICON}
                  View on GitHub
                </a>
                <Link
                  href="/showcase"
                  className="inline-flex items-center gap-2 px-4 py-2.5 border border-line2 rounded-md text-[0.875rem] text-ink2 hover:border-line3 hover:text-ink transition-colors"
                >
                  See it in action
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
                    <path d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                </Link>
              </div>
            </div>

            <Terminal label="zsh — diskwhy scan" bodyHtml={HERO_TERMINAL} />
          </div>
        </div>
      </section>

      {/* HOW IT WORKS */}
      <section className="py-24 border-t border-line">
        <div className="max-w-[1160px] mx-auto px-6">
          <Label>how it works</Label>
          <h2 className="text-[clamp(1.625rem,3vw,2.375rem)] font-semibold tracking-[-0.035em] text-white leading-tight mb-3.5">
            Three commands.
          </h2>
          <p className="text-[0.9375rem] text-ink2 leading-relaxed max-w-[520px]">
            No config required. Scan to see what&apos;s there. Review with <code className="font-mono text-[.85em] text-ink">--dry-run</code>. Remove what you approved.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 mt-14 border border-line rounded-lg overflow-hidden">
            {[
              {
                n: "01", h: "Scan", cmd: "diskwhy scan",
                p: <>Walks your home directory and catalogs space hogs by category. Add <C>--deep</C> to include system dirs. Add <C>--path /dir</C> to target a specific directory.</>,
              },
              {
                n: "02", h: "Review", cmd: "diskwhy clean --all --dry-run",
                p: <>Preview exactly what would be deleted — size, category, and staleness for each item. Items modified in the last 7 days are marked <C>active</C> and always skipped.</>,
              },
              {
                n: "03", h: "Clean", cmd: "diskwhy clean --all --yes",
                p: <>Remove what you approved. Pass <C>--trash</C> to move to Trash instead of permanent delete. Git repos run <C>git gc</C> — the repo stays intact.</>,
              },
            ].map(({ n, h, cmd, p }, i) => (
              <div key={n} className={`bg-panel px-7 py-8 ${i > 0 ? "border-t md:border-t-0 md:border-l border-line" : ""}`}>
                <span className="block font-mono text-[0.7rem] text-brand tracking-[.08em] mb-3.5">{n}</span>
                <div className="text-[0.9375rem] font-semibold text-white tracking-tight mb-2.5">{h}</div>
                <div className="inline-block bg-panel2 border border-line rounded px-2.5 py-1 font-mono text-[0.75rem] text-brand mb-3">{cmd}</div>
                <p className="text-[0.84375rem] text-ink2 leading-relaxed">{p}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CATEGORIES */}
      <section className="py-24 border-t border-line" id="what-it-finds">
        <div className="max-w-[1160px] mx-auto px-6">
          <Label>categories</Label>
          <h2 className="text-[clamp(1.625rem,3vw,2.375rem)] font-semibold tracking-[-0.035em] text-white leading-tight mb-3.5">
            Knows what to look for.
          </h2>
          <p className="text-[0.9375rem] text-ink2 leading-relaxed max-w-[520px]">
            Most disk scanners show you file sizes. diskwhy knows what node_modules is, what Xcode DerivedData is, and why your ~/.npm cache is 4 GB after six months of inactivity.
          </p>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-2.5 mt-11">
            {CATEGORIES.map(({ name, desc }) => (
              <div key={name} className="bg-panel border border-line rounded-md px-4 py-4 hover:border-line2 transition-colors">
                <div className="font-mono text-[0.8rem] text-brand mb-1">{name}</div>
                <div className="text-[0.8125rem] text-ink2 leading-snug">{desc}</div>
              </div>
            ))}
          </div>

          <div className="mt-14">
            <Label>staleness scores</Label>
            <p className="text-[0.9375rem] text-ink2 leading-relaxed max-w-[520px] mt-2">
              Every item gets a staleness score from access time, mtime, or sentinel files (package.json, go.mod, Cargo.toml). Scores are frozen at scan time — the clean phase never re-derives them.
            </p>
            <div className="flex flex-wrap gap-2.5 mt-10">
              {STALENESS.map(({ color, name, desc }) => (
                <div key={name} className="inline-flex items-center gap-2 bg-panel border border-line rounded-md px-3.5 py-2.5">
                  <span className="w-2 h-2 rounded-full shrink-0" style={{ background: color }} />
                  <span className="font-mono text-[0.78rem] text-ink mr-1">{name}</span>
                  <span className="text-ink2 text-[0.8125rem]">{desc}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* FEATURES */}
      <section className="py-24 border-t border-line" id="features">
        <div className="max-w-[1160px] mx-auto px-6">
          <Label>features</Label>
          <h2 className="text-[clamp(1.625rem,3vw,2.375rem)] font-semibold tracking-[-0.035em] text-white leading-tight mb-3.5">
            Built the right way.
          </h2>
          <p className="text-[0.9375rem] text-ink2 leading-relaxed max-w-[520px]">
            No shell-outs. No brittle heuristics. No silent data loss. Engineered to run safely on real developer machines.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-px mt-14 border border-line rounded-lg overflow-hidden bg-line">
            {FEATURES.map(({ icon, title, body }) => (
              <div key={title} className="bg-panel px-7 py-8 hover:bg-panel2 transition-colors">
                <div className="w-9 h-9 bg-brand/10 border border-brand/20 rounded-lg flex items-center justify-center mb-3.5 text-brand">
                  {icon}
                </div>
                <div className="text-[0.9375rem] font-semibold text-white tracking-tight mb-2">{title}</div>
                <p className="text-[0.84375rem] text-ink2 leading-relaxed">{body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-20 border-t border-line">
        <div className="max-w-[1160px] mx-auto px-6">
          <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-12 flex-wrap">
            <div>
              <h2 className="text-[1.625rem] font-semibold tracking-[-0.035em] text-white mb-1.5">Start in 30 seconds.</h2>
              <p className="text-[0.9rem] text-ink2">Requires Go 1.21+. Works on macOS and Linux.</p>
            </div>
            <div className="flex items-center gap-2.5 flex-wrap">
              <div className="min-w-[340px]"><InstallStrip /></div>
              <Link
                href="/showcase"
                className="inline-flex items-center gap-2 px-4 py-2.5 border border-line2 rounded-md text-[0.875rem] text-ink2 hover:border-line3 hover:text-ink transition-colors"
              >
                See examples
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
                  <path d="M5 12h14M12 5l7 7-7 7" />
                </svg>
              </Link>
            </div>
          </div>
        </div>
      </section>

      <Footer />
    </>
  );
}
