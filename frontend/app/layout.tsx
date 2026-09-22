import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "DOAP",
  description: "DirectOrder Agentic Platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="es">
      <body className="min-h-screen bg-zinc-50 text-zinc-900 antialiased">
        {children}
      </body>
    </html>
  );
}
