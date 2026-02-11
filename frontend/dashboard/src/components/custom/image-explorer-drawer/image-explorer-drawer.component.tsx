import { useEffect, useState } from "react"
import { Copy, Tag, Calendar, Database, HardDrive, Hash } from "lucide-react"
import {
  Sheet,
  SheetContent,
  SheetTitle,
} from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { DottedBackground } from "@/components/custom/header/dotted-background";
import {
  listDockerHubTags,
  type DockerHubTag,
} from "@/api/integration.api"

interface UnifiedImageItem {
  id: string
  name: string
  registry: "github" | "docker"
  description?: string
  tags: string[]
  updatedAt: string | Date
  url?: string
  stars?: number
  pulls?: number
  private?: boolean
}

interface ImageExplorerDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: UnifiedImageItem | null
}

export function ImageExplorerDrawer({
  open,
  onOpenChange,
  item
}: ImageExplorerDrawerProps) {
  const [dockerTags, setDockerTags] = useState<DockerHubTag[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (open && item?.registry === 'docker') {
      fetchDockerTags()
    } else {
      setDockerTags([])
      setError(null)
    }
  }, [open, item])

  const fetchDockerTags = async () => {
    if (!item) return
    setLoading(true)
    setError(null)
    try {
      const [namespace, repo] = item.name.split('/')
      // If name doesn't have slash (library images), handle appropriately
      const ns = repo ? namespace : 'library'
      const r = repo || namespace

      const tags = await listDockerHubTags(ns, r)
      setDockerTags(tags)
    } catch (err) {
      setError("Failed to load tags")
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  // Helper to format bytes
  const formatBytes = (bytes: number, decimals = 2) => {
    if (!+bytes) return '0 Bytes'
    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange} modal={false}>
      <SheetContent overlay={false} className="sm:max-w-[800px] w-[800px] p-0 gap-0 border-l shadow-2xl flex flex-col h-full">

        {/* Header - Matching main Header design */}
        <div className="h-14.2 border-b">
          <DottedBackground className="h-14 px-6 items-center bg-transparent" y={6}>
            <div className="flex items-center gap-3 w-full">
              <div className="flex h-8 w-8 items-center justify-center rounded-md border bg-background/50 backdrop-blur-sm">
                <Database className="h-4 w-4 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <SheetTitle className="text-base font-medium truncate">
                  {item?.name || "Image Explorer"}
                </SheetTitle>
              </div>
              {item?.registry && (
                <Badge variant="outline" className="text-xs uppercase tracking-wider opacity-70">
                  {item.registry}
                </Badge>
              )}
            </div>
          </DottedBackground>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6 [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-accent/30 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb:hover]:bg-muted-background/50">

          {item && (
            <>
              {/* Metadata Section */}
              <div className="space-y-4">
                <div>
                  <h3 className="text-sm font-medium mb-1">Description</h3>
                  <p className="text-sm text-muted-foreground leading-relaxed">
                    {item.description || "No description available."}
                  </p>
                </div>

                <div className="flex flex-wrap gap-4 text-xs text-muted-foreground border p-3 rounded-lg bg-muted/5">
                  <div className="flex items-center gap-1.5">
                    <Calendar className="h-3.5 w-3.5" />
                    <span>Updated {new Date(item.updatedAt).toLocaleDateString()}</span>
                  </div>
                  {item.stars !== undefined && (
                    <div className="flex items-center gap-1.5">
                      <span>★</span>
                      <span>{item.stars} stars</span>
                    </div>
                  )}
                  {item.pulls !== undefined && (
                    <div className="flex items-center gap-1.5">
                      <span>↓</span>
                      <span>{new Intl.NumberFormat('en', { notation: "compact" }).format(item.pulls)} pulls</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Tags List */}
              <div>
                <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
                  <Tag className="h-4 w-4" />
                  Available Tags
                </h3>

                {loading ? (
                  <div className="space-y-3">
                    {[1, 2, 3].map(i => (
                      <div key={i} className="h-16 animate-pulse rounded-lg bg-muted/50" />
                    ))}
                  </div>
                ) : error ? (
                  <div className="text-sm text-destructive p-3 rounded border border-destructive/20 bg-destructive/5">
                    {error}
                  </div>
                ) : (
                  <div className="space-y-3">
                    {/* Docker Tags */}
                    {item.registry === 'docker' && dockerTags.map((tag) => (
                      <div key={tag.name} className="group relative flex flex-col gap-2  border p-3 hover:bg-muted/30 transition-colors">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <Badge variant="secondary" className="font-mono text-xs rounded-none">
                              {tag.name}
                            </Badge>
                          </div>
                          <span className="text-xs text-muted-foreground">
                            {new Date(tag.last_updated).toLocaleDateString()}
                          </span>
                        </div>

                        <div className="flex items-center justify-between text-xs text-muted-foreground">
                          <div className="flex items-center gap-4">
                            <span className="flex items-center gap-1">
                              <HardDrive className="h-3 w-3" />
                              {formatBytes(tag.full_size)}
                            </span>
                            <span className="flex items-center gap-1 relative pl-4 before:absolute before:left-0 before:top-1/2 before:-translate-y-1/2 before:h-3 before:w-px before:bg-border">
                              <Hash className="h-3 w-3" />
                              <span className="font-mono truncate min-w-xl" title={tag.digest}>
                                {tag.digest}
                              </span>
                            </span>
                          </div>

                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={() => navigator.clipboard.writeText(`docker pull ${item.name}:${tag.name}`)}
                            title="Copy pull command"
                          >
                            <Copy className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </div>
                    ))}

                    {/* GitHub Tags (Simple list since API provides less detail in current endpoint) */}
                    {item.registry === 'github' && (
                      item.tags.length > 0 ? (
                        item.tags.map(tag => (
                          <div key={tag} className="flex items-center justify-between border p-3 hover:bg-muted/30 transition-colors">
                            <Badge variant="outline" className="font-mono rounded-none text-xs">
                              {tag}
                            </Badge>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-6 w-6"
                              onClick={() => navigator.clipboard.writeText(`docker pull ghcr.io/${item.name.toLowerCase()}:${tag}`)}
                              title="Copy pull command"
                            >
                              <Copy className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        ))
                      ) : (
                        <div className="text-sm text-muted-foreground italic p-4 text-center">
                          No tags found for this image.
                        </div>
                      )
                    )}
                  </div>
                )}
              </div>
            </>
          )}

        </div>
      </SheetContent>
    </Sheet>
  )
}
