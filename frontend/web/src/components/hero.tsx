"use client";

import { ArrowUpRight, Github } from "lucide-react";
import { geistMono } from "../app/fonts";
import { motion } from "framer-motion";
import { siGithub } from "simple-icons";

export function Hero() {
  return (
    <motion.div
      className="border-b-2 border-dotted border-border"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, ease: "easeOut" }}
    >
      <div className="flex flex-col lg:flex-row items-stretch justify-between text-left">
        {/* Left: Text content pinned to bottom */}
        <motion.div
          className="flex flex-col justify-end flex-1 px-4 py-8"
          initial={{ opacity: 0, x: -50 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <motion.a
            href="https://github.com/siddhantprateek/reefline"
            target="_blank"
            rel="noopener noreferrer"
            className={`text-md text-muted-foreground mb-4 hover:text-foreground transition-colors flex items-center gap-2 ${geistMono.className}`}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.4 }}
            whileHover={{ scale: 1.05, transition: { duration: 0.2 } }}
          >
            <svg width={16} height={16} viewBox="0 0 24 24" fill="currentColor"><path d={siGithub.path} /></svg>
            siddhantprateek/reefline <ArrowUpRight size={18} />
          </motion.a>

          <motion.h1
            className="text-4xl md:text-6xl font-light tracking-tighter mb-2 font-sans text-foreground"
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.5 }}
          >
            Scan. Harden. Ship 🚀.
          </motion.h1>

          <motion.p
            className={`text-muted-foreground max-w-2xl my-5 text-sm md:text-lg leading-relaxed ${geistMono.className}`}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.9 }}
          >
            Container image hygiene and runtime security for modern Kubernetes. Scan for vulnerabilities, CIS Docker Benchmark compliance, layer efficiency, and get optimization recommendations.
          </motion.p>

          <motion.div
            className="flex gap-4"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 1.0 }}
          >
            <motion.a
              href="https://github.com/siddhantprateek/reefline"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 py-2 px-4 bg-foreground text-background text-sm hover:opacity-90 transition-opacity"
              whileHover={{ scale: 1.05, y: -1 }}
              whileTap={{ scale: 0.95 }}
            >
              <Github className="h-4 w-4" />
              Star us on GitHub
            </motion.a>
          </motion.div>
        </motion.div>

        {/* Right: Looping video */}
        <motion.div
          className="w-full lg:w-[480px] flex-shrink-0 overflow-hidden border-l-2 border-dotted border-border"
          initial={{ opacity: 0, x: 50 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.8, delay: 0.3 }}
        >
          <video
            src="/reef.mp4"
            autoPlay
            loop
            muted
            playsInline
            className="w-full h-full object-cover"
          />
        </motion.div>
      </div>
    </motion.div>
  );
}
