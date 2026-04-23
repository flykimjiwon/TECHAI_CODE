import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "SuperSOL - 신한 슈퍼솔",
  description: "신한은행 슈퍼솔 데모 앱",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko">
      <body className="antialiased">{children}</body>
    </html>
  );
}
