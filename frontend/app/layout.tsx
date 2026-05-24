import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import NextTopLoader from "nextjs-toploader";
import { cn } from "@/lib/utils";
import { Providers } from "@/components/providers";
import "./globals.css";

const inter = Inter({ subsets: ["latin"], variable: "--font-sans" });
const jetbrainsMono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono" });

export const metadata: Metadata = {
  title: "TaskFlow",
  description: "Collaborative task management",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={cn("font-sans", inter.variable, jetbrainsMono.variable)}>
      <body>
        {/* Top-bar progress for every App Router navigation, including
            router.push from the login form. Default height is 3px which
            is hard to spot on lighter pages — bumping to 4px and adding
            a subtle shadow makes the cue actually visible. The corner
            spinner stays on so a slow redirect or destination data load
            still has *some* indication while the bar is at 99%. */}
        <NextTopLoader
          color="#3b82f6"
          height={4}
          shadow="0 0 10px #3b82f6,0 0 5px #3b82f6"
          showSpinner
          speed={300}
          crawlSpeed={200}
        />
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
