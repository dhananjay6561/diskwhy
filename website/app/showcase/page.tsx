import Link from "next/link";
import Nav from "@/components/Nav";
import Footer from "@/components/Footer";
import Terminal from "@/components/Terminal";

/* ── terminal body HTML strings ── */

const SCAN_BODY = `<span class="tp">~</span> <span class="tc">diskwhy scan</span>

<span class="tsi">  Scanning /Users/dj (quick) ...</span>

<span class="tsi">  macOS / Macintosh HD  ·  499.9 GB total  ·  312.3 GB used  ·  187.6 GB free</span>

<span class="th">  FOUND 4.2 GB across 8 items</span>
<span class="ts">  ─────────────────────────────────────────────────────────────────</span>
<span class="tsi">   1  </span><span class="tcat">node_modules  </span><span class="tsi">~/old-app/node_modules              </span><span class="tsz">1.8 GB  </span><span class="tu">unused</span>
<span class="tsi">   2  </span><span class="tcat">docker        </span><span class="tsi">3 dangling images · 2 unused volumes </span><span class="tsz">923 MB  </span><span class="tz">—</span>
<span class="tsi">   3  </span><span class="tcat">xcode_derived </span><span class="tsi">~/Library/Developer/Xcode/DerivedData</span><span class="tsz">631 MB  </span><span class="tst">stale</span>
<span class="tsi">   4  </span><span class="tcat">node_modules  </span><span class="tsi">~/blog/node_modules                  </span><span class="tsz">412 MB  </span><span class="tr">recent</span>
<span class="tsi">   5  </span><span class="tcat">pycache       </span><span class="tsi">~/scripts/__pycache__                </span><span class="tsz">248 MB  </span><span class="tst">stale</span>
<span class="tsi">   6  </span><span class="tcat">brew_cache    </span><span class="tsi">~/Library/Caches/Homebrew            </span><span class="tsz">178 MB  </span><span class="tu">unused</span>
<span class="tsi">   7  </span><span class="tcat">git_objects   </span><span class="tsi">~/monorepo/.git                      </span><span class="tsz"> 94 MB  </span><span class="tst">stale</span>
<span class="tsi">   8  </span><span class="tcat">logs          </span><span class="tsi">~/Library/Logs                       </span><span class="tsz"> 81 MB  </span><span class="tst">stale</span>
<span class="ts">  ─────────────────────────────────────────────────────────────────</span>

<span class="tsi">  Run: </span><span class="tc">diskwhy clean --all --dry-run</span>

<span class="tp">~</span> <span class="cursor">▋</span>`;

const DRY_RUN_BODY = `<span class="tp">~</span> <span class="tc">diskwhy clean <span class="tf">--all --dry-run</span></span>

<span class="tsi">  Scanning (deep) ...</span>

<span class="th">  CLEAN PREVIEW — no changes will be made</span>
<span class="ts">  ─────────────────────────────────────────────────────</span>
<span class="tok">  ✓</span> <span class="tsi">dry-run  </span><span class="tcat">node_modules  </span><span class="tsi">~/old-app/node_modules   </span><span class="tsz">1.8 GB</span>
<span class="tok">  ✓</span> <span class="tsi">dry-run  </span><span class="tcat">xcode_derived </span><span class="tsi">~/Library/Developer/Xcode</span><span class="tsz">631 MB</span>
<span class="tok">  ✓</span> <span class="tsi">dry-run  </span><span class="tcat">pycache       </span><span class="tsi">~/scripts/__pycache__    </span><span class="tsz">248 MB</span>
<span class="tok">  ✓</span> <span class="tsi">dry-run  </span><span class="tcat">brew_cache    </span><span class="tsi">~/Library/Caches/Homebrew</span><span class="tsz">178 MB</span>
<span class="tgc">  ✓</span> <span class="tsi">dry-run  </span><span class="tcat">git_objects   </span><span class="tsi">~/monorepo/.git          </span><span class="tsz"> 94 MB</span><span class="tgc"> → gc</span>
<span class="tok">  ✓</span> <span class="tsi">dry-run  </span><span class="tcat">logs          </span><span class="tsi">~/Library/Logs           </span><span class="tsz"> 81 MB</span>
<span class="tsk">  ✗</span> <span class="tsi">skipped  </span><span class="tcat">node_modules  </span><span class="tsi">~/blog/node_modules      </span><span class="tsz">412 MB</span><span class="tsk"> active</span>
<span class="ts">  ─────────────────────────────────────────────────────</span>

<span class="tsi">  Recoverable: </span><span class="tsz">3.0 GB  </span><span class="tsi">|  1 item skipped (active)</span>
<span class="tsi">  Docker: 3 images + 2 volumes = 923 MB (included)</span>

<span class="tsi">  Run without </span><span class="tf">--dry-run</span><span class="tsi"> to apply.</span>

<span class="tp">~</span> <span class="cursor">▋</span>`;

const CLEAN_BODY = `<span class="tp">~</span> <span class="tc">diskwhy clean <span class="tf">--all --yes</span></span>

<span class="tsi">  Scanning (deep) ...</span>

<span class="th">  CLEANING 6 items · 3.0 GB recoverable</span>
<span class="ts">  ─────────────────────────────────────────────────────</span>
<span class="tok">  ✓ deleted  </span><span class="tcat">node_modules  </span><span class="tsi">~/old-app/node_modules   </span><span class="tsz">1.8 GB</span>
<span class="tok">  ✓ deleted  </span><span class="tcat">xcode_derived </span><span class="tsi">~/Library/Developer/Xcode</span><span class="tsz">631 MB</span>
<span class="tok">  ✓ deleted  </span><span class="tcat">pycache       </span><span class="tsi">~/scripts/__pycache__    </span><span class="tsz">248 MB</span>
<span class="tok">  ✓ deleted  </span><span class="tcat">brew_cache    </span><span class="tsi">~/Library/Caches/Homebrew</span><span class="tsz">178 MB</span>
<span class="tgc">  ✓ gc run   </span><span class="tcat">git_objects   </span><span class="tsi">~/monorepo/.git          </span><span class="tsz"> 94 MB</span>
<span class="tok">  ✓ deleted  </span><span class="tcat">logs          </span><span class="tsi">~/Library/Logs           </span><span class="tsz"> 81 MB</span>
<span class="tsk">  ✗ skipped  </span><span class="tcat">node_modules  </span><span class="tsi">~/blog/node_modules      </span><span class="tsz">412 MB</span><span class="tsk"> active</span>
<span class="ts">  ─────────────────────────────────────────────────────</span>
<span class="tsi">  Docker pruned: 3 images + 2 volumes  </span><span class="tsz">923 MB</span>
<span class="ts">  ─────────────────────────────────────────────────────</span>

<span class="tok">  Freed: 3.9 GB</span><span class="tsi"> in 6 operations (1 skipped)</span>

<span class="tp">~</span> <span class="cursor">▋</span>`;

const SHELL_BODY = `<span class="tp">~</span> <span class="tc">diskwhy</span>

<span class="tok"> /$$$$$$  /$$$$$$  /$$$$$$  /$$   /$$
| $$__  $$|_  $$_/ /$$__  $$| $$  /$$/
| $$  \\ $$  | $$  | $$  \\__/| $$ /$$/
| $$$$$$$/ /$$$$$$|  $$$$$$/| $$$$/
|_______/ |______/ \\______/ |___/</span>

<span class="tok"> /$$      /$$ /$$   /$$ /$$     /$$
| $$  /$ | $$| $$  | $$|  $$   /$$/
| $$ /$$$| $$| $$  | $$ \\  $$ /$$/
|__/     \\__/|__/  |__/    |__/</span>

<span class="tsi">  Your disk is full. But why?
  Made by DJ
</span>
<span class="tsi">  Commands:
    /scan [--deep] [--path &lt;dir&gt;]
    /clean [--all|--node|--cache|--git|--logs] [--dry-run] [--trash] [--yes]
    /version    /help    /home    /clear    /exit
</span>
<span class="tcat">diskwhy&gt;</span> <span class="tc">/scan --path ~/projects/webapp</span>

<span class="tsi">  Scanning ~/projects/webapp ...</span>

<span class="th">  FOUND 2.1 GB across 3 items</span>
<span class="ts">  ─────────────────────────────────────────</span>
<span class="tsi">  1  </span><span class="tcat">node_modules  </span><span class="tsi">node_modules  </span><span class="tsz">1.9 GB  </span><span class="tst">stale</span>
<span class="tsi">  2  </span><span class="tcat">pycache       </span><span class="tsi">__pycache__   </span><span class="tsz"> 89 MB  </span><span class="tst">stale</span>
<span class="tsi">  3  </span><span class="tcat">git_objects   </span><span class="tsi">.git          </span><span class="tsz"> 94 MB  </span><span class="tst">stale</span>

<span class="tcat">diskwhy&gt;</span> <span class="tc">/clean <span class="tf">--node --cache --yes</span></span>

<span class="tok">  ✓ deleted  </span><span class="tcat">node_modules  </span><span class="tsz">1.9 GB</span>
<span class="tok">  ✓ deleted  </span><span class="tcat">pycache       </span><span class="tsz"> 89 MB</span>

<span class="tok">  Freed: 1.99 GB</span>

<span class="tcat">diskwhy&gt;</span> <span class="tc">/exit</span>

<span class="tok">  bye</span>

<span class="tp">~</span> <span class="cursor">▋</span>`;

const JQ_BODY = `<span class="tp">~</span> <span class="tc">diskwhy scan --json \\
    | jq <span class="tf">'.items | sort_by(-.size_bytes) | .[0:3]'</span></span>

<span class="tsi">[
  {
    "path":            "~/old-app/node_modules",
    "category":        "node_modules",
    "size_bytes":      1879048192,
    "staleness_score": "unused"
  },
  ...
]</span>

<span class="tp">~</span> <span class="tc">diskwhy clean --all --json \\
    | jq <span class="tf">'.summary.freed_bytes / 1073741824 | floor'</span></span>

<span class="tsi">3</span>`;

const SCAN_JSON = `{
  <span class="jk">"schema_version"</span>: <span class="jn">1</span>,
  <span class="jk">"scanned_at"</span>:     <span class="jv">"2026-01-15T10:30:42Z"</span>,
  <span class="jk">"scan_mode"</span>:      <span class="jv">"quick"</span>,
  <span class="jk">"header"</span>:         <span class="jv">"[macOS / Macintosh HD]"</span>,
  <span class="jk">"disk"</span>: {
    <span class="jk">"total_bytes"</span>: <span class="jn">499963174912</span>,
    <span class="jk">"used_bytes"</span>:  <span class="jn">312345678901</span>,
    <span class="jk">"free_bytes"</span>:  <span class="jn">187617496011</span>,
    <span class="jk">"mount"</span>:       <span class="jv">"/"</span>
  },
  <span class="jk">"items"</span>: [
    {
      <span class="jk">"path"</span>:             <span class="jv">"/Users/dj/old-app/node_modules"</span>,
      <span class="jk">"category"</span>:         <span class="jv">"node_modules"</span>,
      <span class="jk">"size_bytes"</span>:       <span class="jn">1879048192</span>,
      <span class="jk">"staleness_score"</span>:  <span class="jv">"unused"</span>,
      <span class="jk">"staleness_source"</span>: <span class="jv">"atime"</span>
    },
    <span class="jp">// ... more items</span>
  ],
  <span class="jk">"summary"</span>: {
    <span class="jk">"total_items"</span>: <span class="jn">8</span>,
    <span class="jk">"total_bytes"</span>: <span class="jn">4509715456</span>,
    <span class="jk">"elapsed_ms"</span>:  <span class="jn">834</span>
  }
}`;

const CLEAN_JSON = `{
  <span class="jk">"schema_version"</span>: <span class="jn">1</span>,
  <span class="jk">"cleaned_at"</span>:     <span class="jv">"2026-01-15T10:31:18Z"</span>,
  <span class="jk">"dry_run"</span>:        <span class="jb">false</span>,
  <span class="jk">"use_trash"</span>:      <span class="jb">false</span>,
  <span class="jk">"results"</span>: [
    {
      <span class="jk">"path"</span>:       <span class="jv">"/Users/dj/old-app/node_modules"</span>,
      <span class="jk">"category"</span>:   <span class="jv">"node_modules"</span>,
      <span class="jk">"outcome"</span>:    <span class="jv">"deleted"</span>,
      <span class="jk">"size_bytes"</span>: <span class="jn">1879048192</span>,
      <span class="jk">"error"</span>:      <span class="jv">""</span>
    },
    {
      <span class="jk">"path"</span>:       <span class="jv">"/Users/dj/monorepo/.git"</span>,
      <span class="jk">"category"</span>:   <span class="jv">"git_objects"</span>,
      <span class="jk">"outcome"</span>:    <span class="jv">"gc_run"</span>,
      <span class="jk">"size_bytes"</span>: <span class="jn">98566144</span>,
      <span class="jk">"error"</span>:      <span class="jv">""</span>
    },
    <span class="jp">// ... more results</span>
  ],
  <span class="jk">"docker_freed_bytes"</span>: <span class="jn">967884800</span>,
  <span class="jk">"summary"</span>: {
    <span class="jk">"deleted"</span>:     <span class="jn">5</span>,
    <span class="jk">"trashed"</span>:     <span class="jn">0</span>,
    <span class="jk">"skipped"</span>:     <span class="jn">1</span>,
    <span class="jk">"errors"</span>:      <span class="jn">0</span>,
    <span class="jk">"freed_bytes"</span>: <span class="jn">3145728000</span>
  }
}`;

/* ── helpers ── */

function DemoSection({ num, title, desc, children }: {
  num: string;
  title: string;
  desc: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <section className="py-20 border-b border-line">
      <div className="max-w-[1160px] mx-auto px-6">
        <div className="flex items-baseline gap-4 mb-8 flex-wrap">
          <span className="font-mono text-[0.7rem] text-brand tracking-widest uppercase">{num}</span>
          <h2 className="text-[clamp(1.25rem,2.5vw,1.625rem)] font-semibold tracking-[-0.03em] text-white">{title}</h2>
        </div>
        <p className="text-[0.9375rem] text-ink2 leading-relaxed max-w-[600px] mb-7">{desc}</p>
        {children}
      </div>
    </section>
  );
}

function FlagTable({ rows }: { rows: [string, string][] }) {
  return (
    <table className="w-full border-collapse mt-9 text-[0.84375rem]">
      <thead>
        <tr>
          <th className="text-left font-medium text-[0.75rem] text-ink2 tracking-wider uppercase pb-3 border-b border-line pr-4 w-48">Flag</th>
          <th className="text-left font-medium text-[0.75rem] text-ink2 tracking-wider uppercase pb-3 border-b border-line">Description</th>
        </tr>
      </thead>
      <tbody>
        {rows.map(([flag, desc]) => (
          <tr key={flag}>
            <td className="font-mono text-[0.8rem] text-brand py-3 border-b border-line pr-4 whitespace-nowrap align-top">{flag}</td>
            <td className="text-ink2 py-3 border-b border-line align-top last:border-b-0">{desc}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function JsonBlock({ title, html, badge }: { title: string; html: string; badge?: string }) {
  return (
    <div>
      <div className="font-mono text-[0.7rem] text-ink2 mb-2 tracking-wider uppercase">{title}</div>
      <div className="bg-panel border border-line rounded-lg overflow-hidden">
        <div className="flex items-center justify-between px-4 py-2.5 border-b border-line bg-panel2">
          <div className="flex items-center gap-2.5">
            <div className="flex gap-1.5">
              <div className="w-2.5 h-2.5 rounded-full bg-[#ff5f57]" />
              <div className="w-2.5 h-2.5 rounded-full bg-[#febc2e]" />
              <div className="w-2.5 h-2.5 rounded-full bg-[#28c840]" />
            </div>
            <span className="font-mono text-[0.7rem] text-ink2">{title.replace(" ", "_")}.json</span>
          </div>
          {badge && <span className="font-mono text-[0.7rem] text-brand bg-brand/10 border border-brand/20 rounded px-2 py-0.5">{badge}</span>}
        </div>
        <div
          className="px-6 py-5 font-mono text-[0.775rem] leading-[1.9] text-[#aaa] overflow-x-auto whitespace-pre"
          dangerouslySetInnerHTML={{ __html: html }}
        />
      </div>
    </div>
  );
}

export default function Showcase() {
  return (
    <>
      <Nav active="showcase" />

      {/* PAGE HEADER */}
      <section className="pt-[120px] pb-16 border-b border-line">
        <div className="max-w-[1160px] mx-auto px-6">
          <div className="flex items-end justify-between gap-8 flex-wrap">
            <div>
              <h1 className="text-[clamp(2rem,4vw,3rem)] font-semibold tracking-[-0.04em] text-white leading-tight mb-3">
                See it in action.
              </h1>
              <p className="text-[0.9375rem] text-ink2 max-w-[440px] leading-relaxed">
                Real command output. Realistic data. No invented scenarios — these reflect exactly what diskwhy produces on a typical developer machine.
              </p>
            </div>
            <div className="flex gap-2.5 flex-wrap items-center">
              <Link
                href="/"
                className="inline-flex items-center gap-2 px-4 py-2.5 border border-line2 rounded-md text-[0.875rem] text-ink2 hover:border-line3 hover:text-ink transition-colors"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
                  <path d="M19 12H5M12 19l-7-7 7-7" />
                </svg>
                Back to home
              </Link>
              <a
                href="https://github.com/dhananjay6561/diskwhy"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 px-4 py-2.5 bg-brand border border-brand rounded-md text-[0.875rem] font-medium text-black hover:bg-green-600 hover:border-green-600 transition-colors"
              >
                <svg viewBox="0 0 16 16" fill="currentColor" className="w-3.5 h-3.5 shrink-0">
                  <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
                </svg>
                Install
              </a>
            </div>
          </div>
        </div>
      </section>

      {/* 01 – scan */}
      <DemoSection
        num="01"
        title="diskwhy scan"
        desc="Quick scan of your home directory. Finds space hogs by category and shows a size-sorted list with staleness scores. The disk header shows total capacity and how much is in use."
      >
        <Terminal label="zsh — ~" bodyHtml={SCAN_BODY} />
        <FlagTable rows={[
          ["--path <dir>", 'Scan a specific directory instead of home. Shows scan_mode: "path" in JSON output.'],
          ["--deep", "Also walks /usr/local, /opt/homebrew, /var/cache, and other system directories."],
          ["--json", "Emit structured JSON with schema_version: 1. Safe to pipe to jq or parse in scripts."],
          ["--verbose", "Show per-directory timing, resolved symlink paths, and staleness source."],
        ]} />
      </DemoSection>

      {/* 02 – clean */}
      <DemoSection
        num="02"
        title="diskwhy clean"
        desc={<>Always run <code className="bg-panel border border-line rounded px-1.5 py-px text-[.85em] text-ink">--dry-run</code> first to preview what will be deleted. Active items are unconditionally skipped. When you&apos;re ready, remove <code className="bg-panel border border-line rounded px-1.5 py-px text-[.85em] text-ink">--dry-run</code> and confirm.</>}
      >
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <div className="font-mono text-[0.7rem] text-ink2 mb-2 tracking-wider uppercase">preview with --dry-run</div>
            <Terminal label="diskwhy clean --all --dry-run" bodyHtml={DRY_RUN_BODY} />
          </div>
          <div>
            <div className="font-mono text-[0.7rem] text-ink2 mb-2 tracking-wider uppercase">run it</div>
            <Terminal label="diskwhy clean --all --yes" bodyHtml={CLEAN_BODY} />
          </div>
        </div>

        <div className="mt-8">
          <p className="text-[0.84375rem] text-ink2 mb-3.5">Every item in the output has an outcome. The six possible states:</p>
          <div className="flex flex-wrap gap-2">
            {[
              { label: "dry_run",  color: "text-ink2",   note: "would have been deleted" },
              { label: "deleted",  color: "text-brand",  note: "removed from disk" },
              { label: "gc_run",   color: "text-violet", note: "git gc ran on the repo" },
              { label: "trashed",  color: "text-sky",    note: "moved to system Trash" },
              { label: "skipped",  color: "text-ink3",   note: "active, filtered, or blocklisted" },
              { label: "error",    color: "text-rose",   note: "deletion failed" },
            ].map(({ label, color, note }) => (
              <div key={label} className="inline-flex items-center gap-1.5 bg-panel border border-line rounded-md px-3 py-1.5 text-[0.8125rem]">
                <span className={`font-mono text-[0.78rem] ${color}`}>{label}</span>
                <span className="text-ink2 text-[0.8rem]">{note}</span>
              </div>
            ))}
          </div>
        </div>
      </DemoSection>

      {/* 03 – shell */}
      <DemoSection
        num="03"
        title="diskwhy shell"
        desc={<>Run <code className="bg-panel border border-line rounded px-1.5 py-px text-[.85em] text-ink">diskwhy</code> with no arguments to open the interactive REPL. Slash commands for scan, clean, help, and more. Useful when you want to run multiple operations without retyping the binary.</>}
      >
        <Terminal label="zsh — diskwhy shell" bodyHtml={SHELL_BODY} />
        <FlagTable rows={[
          ["/scan [flags]",   "Run a scan. Accepts the same flags as diskwhy scan."],
          ["/clean [flags]",  "Run clean. Accepts the same flags as diskwhy clean."],
          ["/help",           "Show available commands and examples."],
          ["/home",           "Re-render the branded home screen."],
          ["/version",        "Print the binary version."],
          ["/clear",          "Clear the terminal screen."],
          ["/exit",           "Quit the shell."],
        ]} />
      </DemoSection>

      {/* 04 – JSON */}
      <DemoSection
        num="04"
        title="JSON output"
        desc={<>Both <code className="bg-panel border border-line rounded px-1.5 py-px text-[.85em] text-ink">diskwhy scan</code> and <code className="bg-panel border border-line rounded px-1.5 py-px text-[.85em] text-ink">diskwhy clean</code> support <code className="bg-panel border border-line rounded px-1.5 py-px text-[.85em] text-ink">--json</code>. The output has a stable <code className="bg-panel border border-line rounded px-1.5 py-px text-[.85em] text-ink">schema_version: 1</code> contract. Pipe to jq, parse in CI, or feed into dashboards.</>}
      >
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-start">
          <JsonBlock title="scan_output" html={SCAN_JSON} badge="schema_version: 1" />
          <div className="flex flex-col gap-4">
            <JsonBlock title="clean_output" html={CLEAN_JSON} badge="schema_version: 1" />
            <div className="bg-panel border border-line rounded-md overflow-hidden">
              <div className="px-4 py-2 border-b border-line bg-panel2 font-mono text-[0.7rem] text-ink2 uppercase tracking-wider">jq example</div>
              <Terminal label="pipe to jq" bodyHtml={JQ_BODY} bodyClass="!p-3.5 !text-[0.75rem]" />
            </div>
          </div>
        </div>
      </DemoSection>

      {/* 05 – completion */}
      <section className="py-20">
        <div className="max-w-[1160px] mx-auto px-6">
          <div className="flex items-baseline gap-4 mb-8">
            <span className="font-mono text-[0.7rem] text-brand tracking-widest uppercase">05</span>
            <h2 className="text-[clamp(1.25rem,2.5vw,1.625rem)] font-semibold tracking-[-0.03em] text-white">Shell completion</h2>
          </div>
          <p className="text-[0.9375rem] text-ink2 leading-relaxed max-w-[600px] mb-7">
            Tab completion for all flags and subcommands, generated by cobra&apos;s built-in completion engine. Install once and forget.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {[
              { shell: "bash",       cmd: "diskwhy completion bash > /usr/local/etc/bash_completion.d/diskwhy\nsource ~/.bashrc" },
              { shell: "zsh",        cmd: `diskwhy completion zsh > "\${fpath[1]}/_diskwhy"\nsource ~/.zshrc` },
              { shell: "fish",       cmd: "diskwhy completion fish > ~/.config/fish/completions/diskwhy.fish" },
              { shell: "powershell", cmd: "diskwhy completion powershell | Out-String | Invoke-Expression" },
            ].map(({ shell, cmd }) => (
              <div key={shell} className="bg-panel border border-line rounded-md overflow-hidden">
                <div className="px-3.5 py-2 border-b border-line bg-panel2 font-mono text-[0.72rem] text-ink2">{shell}</div>
                <pre className="px-3.5 py-3 font-mono text-[0.775rem] text-ink2 overflow-x-auto whitespace-pre">
                  {cmd.split("\n").map((line, i) => (
                    <span key={i} className="block">
                      {line.startsWith("diskwhy") || line.startsWith("source") ? (
                        <span className="text-brand">{line}</span>
                      ) : line}
                    </span>
                  ))}
                </pre>
              </div>
            ))}
          </div>
        </div>
      </section>

      <Footer />
    </>
  );
}
