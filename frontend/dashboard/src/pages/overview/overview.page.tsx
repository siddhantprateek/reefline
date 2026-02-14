import { useEffect, useState, useMemo } from "react"
import {
  Loader2,
  ExternalLink,
  Package,
  RefreshCw,
  Cpu,
  Database,
  Check,
  Lock,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { DottedBackground } from "@/components/custom/header/dotted-background"
import { cn } from "@/lib/utils"
import { ImageExplorerDrawer } from "@/components/custom/image-explorer-drawer/image-explorer-drawer.component"

import {
  type GitHubContainerImage,
  type DockerHubRepo,
  type KubernetesContainerImage,
  listGitHubContainerImages,
  listDockerHubRepos,
  listKubernetesImages,
} from "@/api/integration.api"

// --- Types ---

type RegistryType = "github" | "docker" | "kubernetes"
type VisibilityType = "public" | "private"

export interface UnifiedImageItem {
  id: string
  name: string
  registry: RegistryType
  description?: string
  tags: string[]
  updatedAt: string | Date
  url?: string
  stars?: number
  pulls?: number
  private?: boolean
}

// --- Components ---

function FilterSection({
  title,
  icon: Icon,
  children
}: {
  title: string
  icon: React.ElementType
  children: React.ReactNode
}) {
  return (
    <div className="border-b border-border">
      <DottedBackground className="border-b border-border/50 dark:bg-neutral-900/20" cy={10}>
        <div className="flex items-center gap-2 px-4 py-3">
          <Icon className="h-4 w-4 text-muted-foreground" />
          <h3 className="text-sm font-medium tracking-wide">{title}</h3>
        </div>
      </DottedBackground>
      <div className="p-4 space-y-3">
        {children}
      </div>
    </div>
  )
}

function CheckboxItem({
  label,
  checked,
  onChange,
  count
}: {
  label: string
  checked: boolean
  onChange: (checked: boolean) => void
  count?: number
}) {
  return (
    <label className="flex items-center gap-3 cursor-pointer group select-none">
      <div
        className={cn(
          "h-4 w-4 rounded border border-input flex items-center justify-center transition-colors shadow-sm",
          checked
            ? "bg-primary border-primary text-primary-foreground"
            : "bg-transparent group-hover:border-primary/50"
        )}
      >
        {checked && <Check className="h-3 w-3" />}
        <input
          type="checkbox"
          className="sr-only"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
        />
      </div>
      <span className={cn("text-sm transition-colors", checked ? "font-medium text-foreground" : "text-muted-foreground group-hover:text-foreground")}>
        {label}
      </span>
      {count !== undefined && (
        <span className="ml-auto text-xs text-muted-foreground border px-1.5 py-0.5 rounded-sm bg-muted/30">
          {count}
        </span>
      )}
    </label>
  )
}


interface ImageRowProps {
  item: UnifiedImageItem
  onClick: (item: UnifiedImageItem) => void
}

function ImageRow({ item, onClick }: ImageRowProps) {
  const isGithub = item.registry === 'github'
  const isK8s = item.registry === 'kubernetes'

  return (
    <div
      onClick={() => onClick(item)}
      className="group flex items-center justify-between gap-4 border-b last:border-0 border-border p-4 hover:bg-muted/40 transition-colors cursor-pointer"
    >

      {/* Left: Identity */}
      <div className="flex items-start gap-4 min-w-[300px] max-w-[40%]">
        <div className={cn(
          "flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border bg-background",
          isGithub ? "border-neutral-200 dark:border-neutral-800"
            : isK8s ? "border-purple-200 dark:border-purple-900"
              : "border-blue-200 dark:border-blue-900"
        )}>
          <Package className={cn(
            "h-5 w-5",
            isGithub ? "text-neutral-600 dark:text-neutral-400"
              : isK8s ? "text-purple-600 dark:text-purple-400"
                : "text-blue-600 dark:text-blue-400"
          )} />
        </div>

        <div className="space-y-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-medium text-sm truncate" title={item.name}>
              {item.name}
            </h3>
            {item.private && (
              <Badge variant="secondary" className="h-5 px-1.5 text-[10px]">Private</Badge>
            )}
          </div>
          <p className="text-xs text-muted-foreground line-clamp-1" title={item.description}>
            {item.description || (isGithub ? "Container Image" : "Docker Repository")}
          </p>
        </div>
      </div>

      {/* Middle: Tags (Responsive) */}
      <div className="hidden md:flex flex-1 flex-wrap gap-1.5 px-4">
        {(item.tags || []).slice(0, 3).map(tag => (
          <Badge key={tag} variant="outline" className="px-1.5 py-0 text-[10px] h-5 font-normal border-border bg-background">
            {tag}
          </Badge>
        ))}
        {(item.tags || []).length > 3 && (
          <span className="text-[10px] text-muted-foreground pl-1">
            +{item.tags.length - 3}
          </span>
        )}
      </div>

      {/* Right: Meta & Actions */}
      <div className="flex items-center gap-6 shrink-0 text-sm text-muted-foreground">

        <div className="hidden lg:flex items-center gap-4 text-xs">
          {item.stars !== undefined && (
            <div className="flex items-center gap-1 min-w-[3rem]" title="Stars">
              <span className="font-medium text-foreground">{item.stars}</span> ★
            </div>
          )}
          {item.pulls !== undefined && (
            <div className="flex items-center gap-1 min-w-[4rem]" title="Pulls">
              <span className="font-medium text-foreground">
                {new Intl.NumberFormat('en', { notation: "compact" }).format(item.pulls)}
              </span> ↓
            </div>
          )}
        </div>

        <div className="text-xs w-24 text-right">
          {new Date(item.updatedAt).toLocaleDateString()}
        </div>

        <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" asChild>
          <a href={item.url} target="_blank" rel="noopener noreferrer">
            <ExternalLink className="h-4 w-4" />
          </a>
        </Button>
      </div>

    </div>
  )
}


export function OverviewPage() {
  const [ghImages, setGhImages] = useState<GitHubContainerImage[]>([])
  const [dockerRepos, setDockerRepos] = useState<DockerHubRepo[]>([])
  const [k8sImages, setK8sImages] = useState<KubernetesContainerImage[]>([])

  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Filters
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedRegistries, setSelectedRegistries] = useState<RegistryType[]>([]) // Empty means all
  const [selectedVisibilities, setSelectedVisibilities] = useState<VisibilityType[]>([]) // Empty means all

  // Drawer State
  const [selectedImage, setSelectedImage] = useState<UnifiedImageItem | null>(null)
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)

  const handleImageClick = (item: UnifiedImageItem) => {
    setSelectedImage(item)
    setIsDrawerOpen(true)
  }

  // Fetch logic
  const fetchData = async () => {
    try {
      setError(null)
      const [ghData, dockerData, k8sData] = await Promise.allSettled([
        listGitHubContainerImages(),
        listDockerHubRepos(),
        listKubernetesImages(),
      ])

      if (ghData.status === "fulfilled") {
        setGhImages(ghData.value)
      } else {
        console.error("Failed to fetch GitHub images:", ghData.reason)
      }

      if (dockerData.status === "fulfilled") {
        setDockerRepos(dockerData.value)
      } else {
        console.error("Failed to fetch Docker repos:", dockerData.reason)
      }

      if (k8sData.status === "fulfilled") {
        setK8sImages(k8sData.value)
      } else {
        console.warn("Kubernetes not available or not in-cluster:", k8sData.reason)
      }
    } catch (err) {
      setError("Failed to load some data")
      console.error(err)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const handleRefresh = () => {
    setRefreshing(true)
    fetchData()
  }

  // Normalize data
  const unifiedItems: UnifiedImageItem[] = useMemo(() => {
    const ghItems: UnifiedImageItem[] = ghImages.map(img => ({
      id: `gh-${img.id}`,
      name: img.name,
      registry: 'github',
      description: img.package_type, // GHCR doesn't give description in list usually
      tags: img.tags || [],
      updatedAt: new Date(), // GHCR list endpoint doesn't return updated_at easily, simplified
      url: img.html_url,
      private: true // detailed visibility not always in list, assume private for enterprise usually
    }))

    const dockerItems: UnifiedImageItem[] = dockerRepos.map(repo => ({
      id: `dh-${repo.namespace}-${repo.name}`,
      name: `${repo.namespace}/${repo.name}`,
      registry: 'docker',
      description: repo.description,
      tags: [], // Docker Hub list repo endpoint doesn't return tags, requires detailed fetch. Omitted for list view perf.
      updatedAt: repo.last_updated,
      url: `https://hub.docker.com/r/${repo.namespace}/${repo.name}`,
      stars: repo.star_count,
      pulls: repo.pull_count,
      private: repo.is_private
    }))

    const k8sItems: UnifiedImageItem[] = k8sImages.map((img, idx) => ({
      id: `k8s-${img.namespace}-${img.pod_name}-${img.container_name}-${idx}`,
      name: img.image,
      registry: 'kubernetes',
      description: `${img.namespace} / ${img.pod_name}${img.is_init ? " (init)" : ""}`,
      tags: [img.namespace],
      updatedAt: new Date(),
      private: true,
    }))

    return [...ghItems, ...dockerItems, ...k8sItems].sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
  }, [ghImages, dockerRepos, k8sImages])

  // Filtering
  const filteredItems = useMemo(() => {
    return unifiedItems.filter(item => {
      // 1. Search Query
      if (searchQuery && !item.name.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false
      }
      // 2. Registry Filter
      if (selectedRegistries.length > 0 && !selectedRegistries.includes(item.registry)) {
        return false
      }
      // 3. Visibility Filter
      if (selectedVisibilities.length > 0) {
        const isPrivate = item.private === true
        const isPublic = !isPrivate

        const showPrivate = selectedVisibilities.includes('private')
        const showPublic = selectedVisibilities.includes('public')

        if (showPrivate && !showPublic && !isPrivate) return false
        if (showPublic && !showPrivate && !isPublic) return false
      }
      return true
    })
  }, [unifiedItems, searchQuery, selectedRegistries, selectedVisibilities])

  // Stats for sidebar counts
  const ghCount = unifiedItems.filter(i => i.registry === 'github').length
  const dockerCount = unifiedItems.filter(i => i.registry === 'docker').length
  const k8sCount = unifiedItems.filter(i => i.registry === 'kubernetes').length

  const privateCount = unifiedItems.filter(i => i.private === true).length
  const publicCount = unifiedItems.filter(i => !i.private).length

  const handleRegistryToggle = (reg: RegistryType, checked: boolean) => {
    setSelectedRegistries(prev => {
      if (checked) {
        return [...prev, reg]
      } else {
        return prev.filter(r => r !== reg)
      }
    })
  }

  const handleVisibilityToggle = (vis: VisibilityType, checked: boolean) => {
    setSelectedVisibilities(prev => {
      if (checked) {
        return [...prev, vis]
      } else {
        return prev.filter(v => v !== vis)
      }
    })
  }

  const isAllRegistries = selectedRegistries.length === 0

  return (
    <div className="flex flex-col h-[calc(100vh-theme(spacing.16))] w-full"> {/* Adjust height based on header/layout */}

      {/* Page Header */}
      <div className="flex items-center justify-between px-6 py-5 border-b border-border bg-background/50 backdrop-blur-sm sticky top-0 z-10">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Overview</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Browse and manage container images across connected registries.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search images..."
            className="w-64 h-9"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={loading || refreshing}
            className="h-9"
          >
            <RefreshCw
              className={cn("mr-2 h-3.5 w-3.5", refreshing && "animate-spin")}
            />
            {refreshing ? "Refreshing..." : "Refresh"}
          </Button>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">

        {/* Sidebar Filters */}
        <aside className="w-64 border-r border-border bg-card/30 flex flex-col overflow-y-auto [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-accent/30 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb:hover]:bg-muted-background/50">

          <FilterSection title="Registry" icon={Database}>
            <CheckboxItem
              label="All"
              checked={isAllRegistries}
              onChange={(_checked) => setSelectedRegistries([])}
              count={unifiedItems.length}
            />
            <CheckboxItem
              label="GitHub Container Registry"
              checked={selectedRegistries.includes('github')}
              onChange={(c) => handleRegistryToggle('github', c)}
              count={ghCount}
            />
            <CheckboxItem
              label="Docker Hub"
              checked={selectedRegistries.includes('docker')}
              onChange={(c) => handleRegistryToggle('docker', c)}
              count={dockerCount}
            />
            <CheckboxItem
              label="Kubernetes"
              checked={selectedRegistries.includes('kubernetes')}
              onChange={(c) => handleRegistryToggle('kubernetes', c)}
              count={k8sCount}
            />
          </FilterSection>

          <FilterSection title="Visibility" icon={Lock}>
            <CheckboxItem
              label="All"
              checked={selectedVisibilities.length === 0}
              onChange={() => setSelectedVisibilities([])}
              count={unifiedItems.length}
            />
            <CheckboxItem
              label="Private"
              checked={selectedVisibilities.includes('private')}
              onChange={(c) => handleVisibilityToggle('private', c)}
              count={privateCount}
            />
            <CheckboxItem
              label="Public"
              checked={selectedVisibilities.includes('public')}
              onChange={(c) => handleVisibilityToggle('public', c)}
              count={publicCount}
            />
          </FilterSection>

          <FilterSection title="Architecture" icon={Cpu}>
            <div className="px-1 py-2 text-xs text-muted-foreground italic">
              Architecture filtering requires deep inspection of image manifests.
            </div>
            {/* Placeholder for future implementation */}
            <CheckboxItem
              label="All"
              checked={true}
              onChange={() => { }}
            />
            <CheckboxItem
              label="linux/amd64"
              checked={false}
              onChange={() => { }}
              count={0}
            />
            <CheckboxItem
              label="linux/arm64"
              checked={false}
              onChange={() => { }}
              count={0}
            />
          </FilterSection>

        </aside>

        {/* Main List Content */}
        <main className="flex-1 overflow-y-auto [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-accent/30 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb:hover]:bg-muted-background/50">
          {error && (
            <div className="m-6 mb-2 rounded-md bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20 flex items-center gap-2">
              <span className="font-semibold">Error:</span> {error}
            </div>
          )}

          {loading ? (
            <div className="flex h-64 items-center justify-center flex-col gap-4 text-muted-foreground">
              <Loader2 className="h-8 w-8 animate-spin" />
              <p className="text-sm">Loading registries...</p>
            </div>
          ) : filteredItems.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-center">
              <div className="h-16 w-16 rounded-full bg-muted/50 flex items-center justify-center mb-4">
                <Package className="h-8 w-8 text-muted-foreground" />
              </div>
              <h3 className="text-lg font-medium">No images found</h3>
              <p className="text-muted-foreground text-sm max-w-sm mt-2">
                We couldn't find any container images matching your current filters. Try adjusting your search or filters.
              </p>
              {(searchQuery || selectedRegistries.length > 0 || selectedVisibilities.length > 0) && (
                <Button variant="link" onClick={() => { setSearchQuery(""); setSelectedRegistries([]); setSelectedVisibilities([]) }} className="mt-4">
                  Clear all filters
                </Button>
              )}
            </div>
          ) : (
            <div className="flex flex-col">
              {filteredItems.map(item => (
                <ImageRow key={item.id} item={item} onClick={handleImageClick} />
              ))}
            </div>
          )}
        </main>

        <ImageExplorerDrawer
          open={isDrawerOpen}
          onOpenChange={setIsDrawerOpen}
          item={selectedImage}
        />

      </div>
    </div>
  )
}
