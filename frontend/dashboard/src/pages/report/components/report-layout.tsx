interface ReportLayoutProps {
  left: React.ReactNode;
  right: React.ReactNode;
}

export function ReportLayout({ left, right }: ReportLayoutProps) {
  return (
    <div className="flex h-[calc(100vh-8rem)] overflow-hidden">
      {/* Left Panel - Plan/Recommendations */}
      <div className="w-1/2 border-r border-border overflow-y-auto dark:bg-[#151716]">
        {left}
      </div>

      {/* Right Panel - Tabs */}
      <div className="w-1/2 overflow-y-auto">
        {right}
      </div>
    </div>
  );
}
