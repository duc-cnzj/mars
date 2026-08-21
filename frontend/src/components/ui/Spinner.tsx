/** 加载指示器：居中转圈 */
export function Spinner({ className = '' }: { className?: string }) {
  return (
    <div className={`flex items-center justify-center py-10 ${className}`}>
      <span className="h-6 w-6 animate-spin rounded-full border-2 border-line border-t-primary" />
    </div>
  )
}
