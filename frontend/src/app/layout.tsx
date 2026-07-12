import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";
import { ThemeProvider, QueryProvider } from "@/components/providers";
import { Toaster } from "sonner";
import { RealtimeProviders } from "@/components/realtime-providers";

const geistSans = localFont({
  src: "./fonts/GeistVF.woff",
  variable: "--font-geist-sans",
  weight: "100 900",
});
const geistMono = localFont({
  src: "./fonts/GeistMonoVF.woff",
  variable: "--font-geist-mono",
  weight: "100 900",
});

export const metadata: Metadata = {
  title: "Maid Recruitment Platform",
  description: "Track and manage maid recruitment",
  icons: {
    icon: [
      { url: "/branding/logo-light.webp", media: "(prefers-color-scheme: light)" },
      { url: "/branding/logo-dark.webp", media: "(prefers-color-scheme: dark)" },
    ],
    apple: [{ url: "/branding/logo-light.webp" }],
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const apiOrigin = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="dns-prefetch" href={apiOrigin} />
        <link rel="preconnect" href={apiOrigin} crossOrigin="anonymous" />
        <link rel="dns-prefetch" href="https://pub-ebaf17804d5146cd98dcfec2fae780af.r2.dev" />
        <link rel="preconnect" href="https://pub-ebaf17804d5146cd98dcfec2fae780af.r2.dev" crossOrigin="anonymous" />
        <link rel="manifest" href="/manifest.json" />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <script
          dangerouslySetInnerHTML={{
            __html: `
              if ("serviceWorker" in navigator) {
                window.addEventListener("load", () => {
                  navigator.serviceWorker.register("/sw.js").catch(() => {});
                });
              }
            `,
          }}
        />
        <QueryProvider>
          <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            {children}
            <Toaster
              position="top-right"
              duration={4000}
              richColors
              closeButton
            />
            <RealtimeProviders />
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
