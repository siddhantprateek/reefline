import Link from "next/link";
import Image from "next/image";
import { ArrowLeft } from "lucide-react";
import { geistMono } from "@/app/fonts";

export default function NotFound() {
  return (
    <div className="w-full flex-1 flex flex-col min-h-[calc(100vh-200px)]">
      <div className="flex flex-1">
        {/* Left Sidebar */}
        <div className="w-[80px] md:w-[120px] border-r-2 border-dotted border-border flex-shrink-0">
          <div className="sticky top-20 p-4 h-full min-h-[300px] flex items-start justify-center md:justify-start">
            <span
              className="text-sm font-medium text-muted-foreground uppercase tracking-widest"
              style={{ writingMode: 'vertical-rl', textOrientation: 'mixed' }}
            >
              /Not-Found
            </span>
          </div>
        </div>

        {/* Right Content Area */}
        <div className="flex-1 flex flex-col relative w-full overflow-hidden">
          {/* Main Content */}
          <div className="flex-1 flex flex-col items-center justify-center p-6 md:p-10 z-10 pt-20">
            <h1 className="text-4xl md:text-8xl font-light tracking-tighter mb-4">404</h1>
            <h2 className="text-xl md:text-2xl font-light text-muted-foreground mb-8 text-center max-w-md">
              The page you are looking for doesn't exist or has been moved.
            </h2>

            <Link
              href="/"
              className={`group flex items-center gap-2 px-6 py-3 border border-border/50 rounded-full hover:bg-muted/50 transition-colors ${geistMono.className} text-sm`}
            >
              <ArrowLeft size={16} className="group-hover:-translate-x-1 transition-transform" />
              <span>Back to Home</span>
            </Link>
          </div>

          {/* Bottom Image */}
          <div className="absolute inset-x-0 bottom-0 top-1/2 -z-10 opacity-80 mix-blend-screen pointer-events-none flex justify-center">
            <div className="relative w-full max-w-5xl h-full">
              <Image
                src="/not-found.png"
                alt="404 Illustration"
                fill
                className="object-contain object-bottom"
                priority
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
