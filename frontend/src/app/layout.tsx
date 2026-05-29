import type { Metadata } from "next";
import { Fraunces, Noto_Sans, Noto_Sans_Arabic, Noto_Sans_Ethiopic } from "next/font/google";
import "./globals.css";
import { ThemeProvider, QueryProvider } from "@/components/providers";
import { Toaster } from "sonner";

const fraunces = Fraunces({
  subsets: ["latin"],
  variable: "--font-fraunces",
  weight: ["400", "700"],
});
const notoSans = Noto_Sans({
  subsets: ["latin"],
  variable: "--font-noto-sans",
  weight: ["400", "700"],
});
const notoSansArabic = Noto_Sans_Arabic({
  subsets: ["arabic"],
  variable: "--font-noto-sans-arabic",
  weight: ["400", "700"],
});
const notoSansEthiopic = Noto_Sans_Ethiopic({
  subsets: ["ethiopic"],
  variable: "--font-noto-sans-ethiopic",
  weight: ["400", "700"],
});

export const metadata: Metadata = {
  title: "RecruitMatch | Agency Recruitment Workspaces",
  description: "Track candidate sharing, approvals, and recruitment progress across agency workspaces.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${fraunces.variable} ${notoSans.variable} ${notoSansArabic.variable} ${notoSansEthiopic.variable} antialiased`}
      >
        <QueryProvider>
          <ThemeProvider
            attribute="class"
            defaultTheme="dark"
            enableSystem
            disableTransitionOnChange
          >
            {children}
            <Toaster />
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
