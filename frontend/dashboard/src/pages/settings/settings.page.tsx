import { useState } from "react";
import {
  User,
  Bell,
  KeyRound,
  Palette,
  Moon,
  Sun,
  Monitor,
  Copy,
  Eye,
  EyeOff,
  Check,
  RotateCcw,
  ExternalLink,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { useTheme } from "@/components/theme-provider";
import { DottedBackground } from "@/components/custom/header/dotted-background";
import { cn } from "@/lib/utils";

// --- Toggle Switch ---
function Toggle({
  checked,
  onChange,
  id,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  id?: string;
}) {
  return (
    <button
      id={id}
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className={cn(
        "relative inline-flex h-5 w-9 cursor-pointer items-center rounded-full transition-colors duration-200",
        checked ? "bg-primary" : "bg-muted"
      )}
    >
      <span
        className={cn(
          "inline-block h-3.5 w-3.5 rounded-full bg-white shadow-sm transition-transform duration-200",
          checked ? "translate-x-[1.15rem]" : "translate-x-[0.2rem]"
        )}
      />
    </button>
  );
}

// --- Row + Divider ---
function SettingsRow({
  label,
  description,
  action,
}: {
  label: string;
  description?: string;
  action: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between py-3">
      <div>
        <p className="text-sm font-medium">{label}</p>
        {description && (
          <p className="text-xs text-muted-foreground mt-0.5">{description}</p>
        )}
      </div>
      {action}
    </div>
  );
}

// --- Section Nav Item ---
function SectionNavItem({
  icon: Icon,
  label,
  active,
  onClick,
}: {
  icon: React.ElementType;
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "group flex items-center gap-2.5 px-4 py-3 text-sm font-medium transition-all duration-150 w-full text-left cursor-pointer border-b border-border relative overflow-hidden",
        active
          ? "bg-primary/[0.03] text-foreground"
          : "text-muted-foreground hover:text-foreground"
      )}
    >
      <div className="absolute inset-0 bg-gradient-to-t from-primary/5 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
      <div
        className={cn(
          "relative p-1.5 border transition-colors duration-300",
          active
            ? "bg-gradient-to-br from-primary/10 to-primary/5 border-primary/40"
            : "bg-gradient-to-br from-primary/5 to-transparent border-primary/10 group-hover:border-primary/30"
        )}
      >
        <Icon className="h-4 w-4 text-primary" />
      </div>
      <span className="relative">{label}</span>
      {active && (
        <div className="absolute left-0 top-0 bottom-0 w-[2px] bg-primary" />
      )}
    </button>
  );
}

// --- Sections ---
type Section = "profile" | "notifications" | "api" | "appearance";

const sections: { id: Section; label: string; icon: React.ElementType }[] = [
  { id: "profile", label: "Profile", icon: User },
  { id: "notifications", label: "Notifications", icon: Bell },
  { id: "api", label: "API & Security", icon: KeyRound },
  { id: "appearance", label: "Appearance", icon: Palette },
];

// --- Section Title Bar with Dotted Background ---
function SectionTitleBar({
  icon: Icon,
  title,
  badge,
}: {
  icon: React.ElementType;
  title: string;
  badge?: string;
}) {
  return (
    <DottedBackground className="border-b border-border dark:bg-neutral-900/20" cy={10}>
      <div className="flex items-center gap-3 px-5 py-2 w-full ">
        <Icon className="h-4 w-4 text-primary" />
        <span className="text-sm font-medium">{title}</span>
        {badge && (
          <Badge variant="outline" className="text-xs ml-auto">
            {badge}
          </Badge>
        )}
      </div>
    </DottedBackground>
  );
}

// --- Profile Section ---
function ProfileSection() {
  const [name, setName] = useState("Siddhant");
  const [email, setEmail] = useState("siddhant@reefline.io");
  const [org, setOrg] = useState("Reefline Labs");
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div>
      <SectionTitleBar icon={User} title="Profile Information" badge="Account" />
      <div className="p-5 space-y-5">
        <div className="flex items-center gap-4">
          <div className="flex h-14 w-14 items-center justify-center bg-gradient-to-br from-primary/10 to-primary/5 text-primary text-xl font-bold select-none border border-primary/20">
            {name.charAt(0).toUpperCase()}
          </div>
          <div>
            <p className="text-sm font-medium">{name}</p>
            <p className="text-xs text-muted-foreground">{email}</p>
          </div>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="settings-name" className="text-xs">Display Name</Label>
            <Input id="settings-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="settings-email" className="text-xs">Email</Label>
            <Input id="settings-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
          </div>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="settings-org" className="text-xs">Organization</Label>
          <Input id="settings-org" value={org} onChange={(e) => setOrg(e.target.value)} />
        </div>


        <div className="flex justify-end">
          <Button size="sm" onClick={handleSave} className="min-w-[100px]">
            {saved ? (
              <span className="flex items-center gap-1.5">
                <Check className="h-3.5 w-3.5" /> Saved
              </span>
            ) : (
              "Save Changes"
            )}
          </Button>
        </div>
      </div>

      {/* Danger Zone */}
      <div className="border-t border-border">
        <SectionTitleBar icon={User} title="Danger Zone" badge="Destructive" />
        <div className="p-5 bg-destructive/5">
          <div className="flex items-center justify-between  px-4 py-3">
            <div>
              <p className="text-sm font-medium">Delete Account</p>
              <p className="text-xs text-muted-foreground">
                Permanently remove your account and all data.
              </p>
            </div>
            <Button variant="destructive" size="sm">
              Delete
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

// --- Notifications Section ---
function NotificationsSection() {
  const [jobComplete, setJobComplete] = useState(true);
  const [jobFailed, setJobFailed] = useState(true);
  const [integrationAlerts, setIntegrationAlerts] = useState(true);
  const [emailNotifs, setEmailNotifs] = useState(true);
  const [slackNotifs, setSlackNotifs] = useState(false);
  const [weeklyDigest, setWeeklyDigest] = useState(false);

  return (
    <div>
      <SectionTitleBar icon={Bell} title="Job Notifications" badge="Alerts" />
      <div className="px-5 py-2">
        <SettingsRow label="Job completed" description="Notify when a job finishes successfully." action={<Toggle id="notif-job-complete" checked={jobComplete} onChange={setJobComplete} />} />
        <SettingsRow label="Job failed" description="Alert when a job encounters an error." action={<Toggle id="notif-job-failed" checked={jobFailed} onChange={setJobFailed} />} />
        <SettingsRow label="Integration alerts" description="Notify when a connected integration changes." action={<Toggle id="notif-integration" checked={integrationAlerts} onChange={setIntegrationAlerts} />} />
      </div>

      <div className="border-t border-border">
        <SectionTitleBar icon={Bell} title="Delivery Channels" badge="Channels" />
        <div className="px-5 py-2">
          <SettingsRow label="Email notifications" description="Receive alerts to your email address." action={<Toggle id="notif-email" checked={emailNotifs} onChange={setEmailNotifs} />} />
          <SettingsRow
            label="Slack notifications"
            description="Send alerts to a Slack channel."
            action={
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="text-xs">Soon</Badge>
                <Toggle id="notif-slack" checked={slackNotifs} onChange={setSlackNotifs} />
              </div>
            }
          />
          <SettingsRow label="Weekly digest" description="Receive a weekly summary of activity." action={<Toggle id="notif-digest" checked={weeklyDigest} onChange={setWeeklyDigest} />} />
        </div>
      </div>
    </div>
  );
}

// --- API & Security Section ---
function ApiSection() {
  const [showKey, setShowKey] = useState(false);
  const [copied, setCopied] = useState(false);
  const apiKey = "rf_live_a8f3k29d7x1bQ4mNpEz6WtRcYsH0jLuG";

  const handleCopy = () => {
    navigator.clipboard.writeText(apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div>
      <SectionTitleBar icon={KeyRound} title="API Keys" badge="Security" />
      <div className="p-5 space-y-4">
        <div className="space-y-1.5">
          <Label className="text-xs">Live API Key</Label>
          <div className="relative">
            <Input
              readOnly
              value={showKey ? apiKey : "rf_live_••••••••••••••••••••••••"}
              className="pr-20 font-mono text-xs"
            />
            <div className="absolute right-1 top-1/2 -translate-y-1/2 flex items-center gap-0.5">
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setShowKey(!showKey)}>
                {showKey ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
              </Button>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleCopy}>
                {copied ? <Check className="h-3.5 w-3.5 text-green-500" /> : <Copy className="h-3.5 w-3.5" />}
              </Button>
            </div>
          </div>
          <p className="text-xs text-muted-foreground">Created Jan 15, 2026 · Last used 2h ago</p>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <RotateCcw className="mr-1.5 h-3.5 w-3.5" />
            Regenerate
          </Button>
          <Button variant="outline" size="sm">
            <ExternalLink className="mr-1.5 h-3.5 w-3.5" />
            API Docs
          </Button>
        </div>
      </div>

      <div className="border-t border-border">
        <SectionTitleBar icon={KeyRound} title="Security" badge="Protection" />
        <div className="px-5 py-2">
          <SettingsRow label="Two-factor authentication" description="Add an extra layer of security." action={<Button variant="outline" size="sm">Enable</Button>} />
          <SettingsRow label="Active sessions" description="1 active session on this device." action={<Badge variant="secondary" className="text-xs">1 session</Badge>} />
        </div>
      </div>
    </div>
  );
}

// --- Appearance Section ---
function AppearanceSection() {
  const { theme, setTheme } = useTheme();

  const themes: { id: "light" | "dark" | "system"; label: string; icon: React.ElementType }[] = [
    { id: "light", label: "Light", icon: Sun },
    { id: "dark", label: "Dark", icon: Moon },
    { id: "system", label: "System", icon: Monitor },
  ];

  return (
    <div>
      <SectionTitleBar icon={Palette} title="Theme" badge="Display" />
      <div className="p-5 space-y-4">
        <div className="flex gap-2">
          {themes.map((t) => (
            <button
              key={t.id}
              onClick={() => setTheme(t.id)}
              className={cn(
                "group/theme flex flex-1 items-center gap-2 border px-3 py-2.5 cursor-pointer transition-all duration-200 relative overflow-hidden",
                theme === t.id
                  ? "border-primary bg-primary/[0.03]"
                  : "border-border hover:border-primary/40"
              )}
            >
              <div className="absolute inset-0 bg-gradient-to-t from-primary/5 via-transparent to-transparent opacity-0 group-hover/theme:opacity-100 transition-opacity duration-300" />
              <div
                className={cn(
                  "relative flex h-8 w-8 items-center justify-center border transition-colors",
                  theme === t.id
                    ? "bg-gradient-to-br from-primary/10 to-primary/5 border-primary/40"
                    : "bg-gradient-to-br from-primary/5 to-transparent border-primary/10"
                )}
              >
                <t.icon className="h-4 w-4 text-primary" />
              </div>
              <span className="relative text-sm font-medium">{t.label}</span>
              {theme === t.id && (
                <Check className="h-3.5 w-3.5 text-primary ml-auto relative" />
              )}
            </button>
          ))}
        </div>
      </div>

      <div className="border-t border-border">
        <SectionTitleBar icon={Palette} title="Display" badge="Layout" />
        <div className="px-5 py-2">
          <SettingsRow label="Compact mode" description="Reduce spacing for denser layouts." action={<Toggle id="compact-mode" checked={false} onChange={() => { }} />} />
          <SettingsRow label="Show sidebar labels" description="Display text labels next to sidebar icons." action={<Toggle id="sidebar-labels" checked={true} onChange={() => { }} />} />
        </div>
      </div>
    </div>
  );
}

// --- Main Settings Page ---
export function SettingsPage() {
  const [activeSection, setActiveSection] = useState<Section>("profile");

  const renderSection = () => {
    switch (activeSection) {
      case "profile":
        return <ProfileSection />;
      case "notifications":
        return <NotificationsSection />;
      case "api":
        return <ApiSection />;
      case "appearance":
        return <AppearanceSection />;
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 md:p-6">
        <h1 className="text-3xl font-medium tracking-tight bg-gradient-to-br from-foreground to-foreground/70 bg-clip-text">
          Settings
        </h1>
        <p className="text-muted-foreground">
          Manage your account, notifications, and preferences
        </p>
      </div>

      {/* Content — sidebar nav + detail panel */}
      <div className="flex flex-col md:flex-row border-t border-border flex-1">
        {/* Section Navigation */}
        <nav className="md:w-56 shrink-0 border-r border-border">
          {sections.map((section) => (
            <SectionNavItem
              key={section.id}
              icon={section.icon}
              label={section.label}
              active={activeSection === section.id}
              onClick={() => setActiveSection(section.id)}
            />
          ))}
        </nav>

        {/* Section Content */}
        <div className="flex-1 min-w-0">
          {renderSection()}
        </div>
      </div>
    </div>
  );
}
