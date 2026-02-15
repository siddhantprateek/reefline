"use client";

import Link from "next/link";
import { ArrowUpRight, Menu, X, CalendarDays, Linkedin, Twitter, Play } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { useState } from "react";
import { siGithub, siGmail, siGooglescholar } from "simple-icons";

export function Header() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  return (
    <>
      <motion.header
        className="bg-transparent"
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
      >
        <nav className="flex justify-between items-center px-4 md:px-6 py-4">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <Link href="/" className="text-sm flex items-center gap-2">
              <motion.div
                className="w-8 h-8 rounded flex items-center justify-center text-muted-foreground"
                whileHover={{ scale: 1.1, rotate: 5 }}
                transition={{ duration: 0.2 }}
              >
                Reefline.ai
              </motion.div>
            </Link>
          </motion.div>

          {/* Desktop Menu */}
          <motion.div
            className="hidden md:flex items-center gap-8"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <Link
              href="/llms.txt"
              className="flex items-center gap-1 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              llms.txt <ArrowUpRight size={18} />
            </Link>
            <motion.a
              href="https://www.youtube.com/watch?v=rQRcPPCOZ_w"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center w-fit py-1 cursor-pointer px-1 bg-foreground text-background text-sm hover:opacity-90 transition-opacity"
              whileHover={{ scale: 1.05, y: -1 }}
              whileTap={{ scale: 0.95 }}
            >
              <motion.div
                className='bg-muted/20 p-1.5'
                whileHover={{ rotate: 180 }}
                transition={{ duration: 0.3 }}
              >
                <Play size={16} fill="currentColor" />
              </motion.div>
              <span className='px-2'>Watch Demo</span>
            </motion.a>
            <motion.a
              href="https://github.com/siddhantprateek/reefline"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center w-fit py-1 cursor-pointer px-1 bg-foreground text-background text-sm hover:opacity-90 transition-opacity"
              whileHover={{ scale: 1.05, y: -1 }}
              whileTap={{ scale: 0.95 }}
            >
              <motion.div
                className='bg-muted/20 p-1.5'
                whileHover={{ rotate: 180 }}
                transition={{ duration: 0.3 }}
              >
                <svg width={16} height={16} viewBox="0 0 24 24" fill="currentColor"><path d={siGithub.path} /></svg>
              </motion.div>
              <span className='px-2'>GitHub</span>
            </motion.a>
          </motion.div>

          {/* Mobile Menu Button */}
          <motion.button
            className="md:hidden text-muted-foreground hover:text-foreground transition-colors"
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            {isMobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </motion.button>
        </nav>
      </motion.header>

      {/* Mobile Sidebar */}
      <AnimatePresence>
        {isMobileMenuOpen && (
          <>
            {/* Backdrop */}
            <motion.div
              className="fixed inset-0 bg-background/80 backdrop-blur-sm z-40 md:hidden"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setIsMobileMenuOpen(false)}
            />

            {/* Sidebar */}
            <motion.div
              className="fixed top-0 right-0 h-full w-[320px] bg-background border-l-2 border-dotted border-border z-50 md:hidden overflow-y-auto"
              initial={{ x: "100%" }}
              animate={{ x: 0 }}
              exit={{ x: "100%" }}
              transition={{ type: "spring", damping: 25, stiffness: 200 }}
            >
              <div className="p-6 flex flex-col h-full">
                {/* Close button */}
                <div className="flex justify-end mb-8">
                  <button
                    onClick={() => setIsMobileMenuOpen(false)}
                    className="text-muted-foreground hover:text-foreground transition-colors"
                  >
                    <X size={24} />
                  </button>
                </div>

                {/* Menu Items */}
                <div className="flex flex-col gap-6 flex-1">
                  {/* llms.txt */}
                  <Link
                    href="/llms.txt"
                    className="flex items-center gap-2 text-base font-medium text-muted-foreground hover:text-foreground transition-colors border-b border-dotted border-border pb-4"
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    llms.txt <ArrowUpRight size={18} />
                  </Link>

                  {/* Watch Demo */}
                  <motion.a
                    href="https://www.youtube.com/watch?v=rQRcPPCOZ_w"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center w-full py-2 cursor-pointer px-2 bg-foreground text-background text-sm hover:opacity-90 transition-opacity"
                    whileTap={{ scale: 0.95 }}
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    <motion.div
                      className='bg-muted/20 p-1.5'
                      whileHover={{ rotate: 180 }}
                      transition={{ duration: 0.3 }}
                    >
                      <Play size={16} fill="currentColor" />
                    </motion.div>
                    <span className='px-2'>Watch Demo</span>
                  </motion.a>

                  {/* GitHub */}
                  <motion.a
                    href="https://github.com/siddhantprateek/reefline"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center w-full py-2 cursor-pointer px-2 bg-foreground text-background text-sm hover:opacity-90 transition-opacity mb-4"
                    whileTap={{ scale: 0.95 }}
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    <motion.div
                      className='bg-muted/20 p-1.5'
                      whileHover={{ rotate: 180 }}
                      transition={{ duration: 0.3 }}
                    >
                      <svg width={16} height={16} viewBox="0 0 24 24" fill="currentColor"><path d={siGithub.path} /></svg>
                    </motion.div>
                    <span className='px-2'>GitHub</span>
                  </motion.a>
                </div>

                {/* Book 30-min call and Social Icons at bottom */}
                <div className="mt-auto pt-6 border-t border-dotted border-border flex flex-col gap-6">

                  {/* Social Icons */}
                  <div className="flex flex-col gap-4">
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">Connect</p>
                    <div className="flex gap-4">
                      {[
                        { href: "https://www.linkedin.com/in/siddhantprateek/", icon: <Linkedin size={20} />, label: "LinkedIn" },
                        { href: "https://github.com/siddhantprateek", icon: <svg width={20} height={20} viewBox="0 0 24 24" fill="currentColor"><path d={siGithub.path} /></svg>, label: "GitHub" },
                        { href: "mailto:meetsiddhantprateek@gmail.com", icon: <svg width={20} height={20} viewBox="0 0 24 24" fill="currentColor"><path d={siGmail.path} /></svg>, label: "Gmail" },
                        { href: "https://scholar.google.com/citations?user=siddhantprateek", icon: <svg width={20} height={20} viewBox="0 0 24 24" fill="currentColor"><path d={siGooglescholar.path} /></svg>, label: "Scholar" },
                        { href: "https://x.com/siddhantprateek", icon: <Twitter size={20} />, label: "X" }
                      ].map((social) => (
                        <a
                          key={social.href}
                          href={social.href}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-muted-foreground hover:text-foreground transition-colors"
                          onClick={() => setIsMobileMenuOpen(false)}
                        >
                          {social.icon}
                        </a>
                      ))}
                    </div>
                  </div>

                  <motion.a
                    href="https://cal.com/siddhantprateek/chat"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center w-full py-2 cursor-pointer px-2 bg-foreground text-background text-sm hover:opacity-90 transition-opacity"
                    whileTap={{ scale: 0.95 }}
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    <motion.div
                      className='bg-muted/20 p-1.5'
                      whileHover={{ rotate: 180 }}
                      transition={{ duration: 0.3 }}
                    >
                      <CalendarDays className="h-4 w-4" />
                    </motion.div>
                    <span className='px-2'>Book 30-min call</span>
                  </motion.a>


                </div>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
}
