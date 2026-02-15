import type { Metadata } from "next";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { Header } from "@/components/header";
import { Footer } from "@/components/footer";
import { dmSans, geistMono } from "./fonts";


export const metadata: Metadata = {
  title: "Meet Siddhant Prateek",
  description: "Welcome to my corner of the internet.",
  twitter: {
    card: "summary_large_image",
    images: ["/twitter-meta.jpg"],
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${dmSans.variable} ${geistMono.variable} font-sans antialiased min-h-screen flex flex-col`}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="dark"
          enableSystem
          disableTransitionOnChange
        >
          {/* Header with full-width bottom border */}
          <header className="border-b-2 border-dotted border-border">
            <div className="px-2 md:px-6">
              <div className="max-w-7xl mx-auto border-x-2 border-dotted border-border">
                <Header />
              </div>
            </div>
          </header>

          {/* Main content with constrained vertical borders */}
          <div className="flex-grow px-2 md:px-6">
            <main className="max-w-7xl mx-auto border-x-2 border-dotted border-border min-h-[calc(100vh-200px)] flex flex-col">
              {children}
            </main>
          </div>

          {/* Footer with full-width top border */}
          <footer className="border-t-2 border-dotted border-border">
            <div className="px-2 md:px-6">
              <div className="max-w-7xl mx-auto border-x-2 border-dotted border-border">
                <Footer />
              </div>
            </div>
          </footer>
        </ThemeProvider>
      </body>
    </html>
  );
}
