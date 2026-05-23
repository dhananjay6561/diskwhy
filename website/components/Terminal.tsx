interface TerminalProps {
  label?: string;
  bodyHtml: string;
  className?: string;
  bodyClass?: string;
}

export default function Terminal({ label, bodyHtml, className = "", bodyClass = "" }: TerminalProps) {
  return (
    <div className={`bg-[#070707] border border-line rounded-[10px] overflow-hidden shadow-[0_32px_64px_rgba(0,0,0,.5)] ${className}`}>
      <div className="flex items-center px-4 py-3 bg-panel border-b border-line gap-2.5">
        <div className="flex gap-1.5">
          <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
          <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
          <div className="w-3 h-3 rounded-full bg-[#28c840]" />
        </div>
        {label && (
          <div className="flex-1 text-center font-mono text-[0.7rem] text-ink3 select-none">{label}</div>
        )}
      </div>
      <div
        className={`p-6 font-mono text-[0.775rem] leading-[1.85] text-[#999] overflow-x-auto whitespace-pre ${bodyClass}`}
        dangerouslySetInnerHTML={{ __html: bodyHtml }}
      />
    </div>
  );
}
