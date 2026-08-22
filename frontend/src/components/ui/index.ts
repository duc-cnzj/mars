/** 组件库统一出口：保留的自研组件 + shadcn 组件（feature 也可直接 import shadcn 路径） */

// 保留的自研组件
export { Empty } from './Empty'
export { Spinner } from './Spinner'
export { Kbd } from './Kbd'
export { StatusDot } from './StatusDot'
export { ThemeSwitcher } from './ThemeSwitcher'
export { Tag, type Tone } from './Tag'
export { SkeletonGrid } from './SkeletonGrid'
export { SkeletonList } from './SkeletonList'
export { SkeletonDetail } from './SkeletonDetail'
export { SkeletonTabLog } from './SkeletonTabLog'
export { SkeletonTabShell } from './SkeletonTabShell'
export { SkeletonTabEdit } from './SkeletonTabEdit'
export { SegmentedSkeleton } from './SegmentedSkeleton'

// shadcn 组件统一出口
export { Button, buttonVariants } from './shadcn/button'
export { Input } from './shadcn/input'
export { Textarea } from './shadcn/textarea'
export { Label } from './shadcn/label'
export { Switch } from './shadcn/switch'
export { Badge, badgeVariants } from './shadcn/badge'
export { Avatar, AvatarImage, AvatarFallback } from './shadcn/avatar'
export { Skeleton } from './shadcn/skeleton'
export { Progress } from './shadcn/progress'
export { Separator } from './shadcn/separator'
export { ScrollArea, ScrollBar } from './shadcn/scroll-area'
export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from './shadcn/tooltip'
export { Tabs, TabsList, TabsTrigger, TabsContent } from './shadcn/tabs'
export {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
  DialogTrigger,
} from './shadcn/dialog'
export {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from './shadcn/alert-dialog'
export {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from './shadcn/dropdown-menu'
export { Popover, PopoverAnchor, PopoverContent, PopoverTrigger } from './shadcn/popover'
export {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from './shadcn/select'
export {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from './shadcn/sheet'
export {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from './shadcn/pagination'
