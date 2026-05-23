import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "diskwhy — your disk is full. but why?",
  description:
    "diskwhy finds what's consuming your disk — node_modules, Docker images, Python caches, Xcode junk — and removes them safely.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen">{children}</body>
    </html>
  );
}
