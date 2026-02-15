"use client";

import { geistMono } from "../app/fonts";
import { motion } from "framer-motion";
import { ThemeSwitcher } from "./theme-switcher";
import { siGithub } from "simple-icons";

export function Footer() {
  const footerLinks = [
    { href: "https://github.com/siddhantprateek/reefline", label: "GitHub", external: true },
    { href: "mailto:meetsiddhantprateek@gmail.com", label: "Contact", external: false },
    { href: "https://github.com/siddhantprateek/reefline/issues", label: "Report an issue", external: true }
  ];

  return (
    <motion.footer
      className="py-10 md:py-20 px-6 mt-auto"
      initial={{ opacity: 0, y: 50 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-50px" }}
      transition={{ duration: 0.6, ease: "easeOut" }}
    >
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-8 md:gap-12">
        <motion.div
          className="flex flex-col items-start gap-4 md:gap-6 max-w-lg"
          initial={{ opacity: 0, x: -30 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <motion.div
            className={`text-base md:text-lg font-medium text-foreground ${geistMono.className}`}
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <p>Scan. Harden. Ship.</p>
            <motion.p
              className="text-muted-foreground"
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.4 }}
            >
              Open source container security for Kubernetes. Contribute on{' '}<motion.a
                href="https://github.com/siddhantprateek/reefline"
                target="_blank"
                rel="noopener noreferrer"
                whileHover={{ scale: 1.05 }}
                className="hover:text-foreground transition-colors"
              >
                GitHub
              </motion.a>
            </motion.p>
          </motion.div>

          <div className="flex items-center gap-4 flex-wrap">
            <motion.a
              href="https://github.com/siddhantprateek/reefline"
              target="_blank"
              rel="noopener noreferrer"
              className="bg-foreground text-background px-4 py-2 rounded font-medium flex items-center gap-2 hover:opacity-90 transition-opacity text-sm md:text-base"
              initial={{ opacity: 0, scale: 0.9 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: 0.5 }}
              whileHover={{ scale: 1.05, y: -2 }}
              whileTap={{ scale: 0.95 }}
            >
              Star on GitHub
              <motion.div
                whileHover={{ rotate: 180 }}
                transition={{ duration: 0.3 }}
              >
                <svg width={16} height={16} viewBox="0 0 24 24" fill="currentColor"><path d={siGithub.path} /></svg>
              </motion.div>
            </motion.a>
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: 0.6 }}
            >
              <ThemeSwitcher />
            </motion.div>
          </div>
        </motion.div>

        <motion.div
          className={`flex flex-col items-start md:items-end gap-4 md:gap-1 text-left md:text-right ${geistMono.className}`}
          initial={{ opacity: 0, x: 30 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <motion.p
            className="text-muted-foreground mb-2 md:mb-4 text-sm md:text-base"
            initial={{ opacity: 0, y: 10 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.4, delay: 0.3 }}
          >
            © {new Date().getFullYear()} Reefline
          </motion.p>

          <div className="flex flex-col gap-2 md:gap-1 text-muted-foreground text-sm md:text-base">
            {footerLinks.map((link, index) => (
              <motion.a
                key={link.href}
                href={link.href}
                target={link.external ? "_blank" : undefined}
                rel={link.external ? "noopener noreferrer" : undefined}
                className="hover:text-foreground transition-colors"
                initial={{ opacity: 0, x: 20 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.3, delay: 0.4 + index * 0.1 }}
                whileHover={{ scale: 1.05, x: -2 }}
              >
                {link.label}
              </motion.a>
            ))}
          </div>
        </motion.div>
      </div>
    </motion.footer>
  );
}
