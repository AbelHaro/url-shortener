import { useTheme } from "next-themes"
import { Toaster as Sonner, type ToasterProps } from "sonner"
import { CheckCircleIcon, InfoIcon, WarningIcon, XCircleIcon, SpinnerIcon } from "@phosphor-icons/react"

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "system" } = useTheme()

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      icons={{
        success: (
          <CheckCircleIcon className="size-4" />
        ),
        info: (
          <InfoIcon className="size-4" />
        ),
        warning: (
          <WarningIcon className="size-4" />
        ),
        error: (
          <XCircleIcon className="size-4" />
        ),
        loading: (
          <SpinnerIcon className="size-4 animate-spin" />
        ),
      }}
      style={
        {
          "--normal-bg": "var(--popover)",
          "--normal-text": "var(--popover-foreground)",
          "--normal-border": "var(--border)",
          "--border-radius": "var(--radius)",
        } as React.CSSProperties
      }
      toastOptions={{
        classNames: {
          toast: "!shadow-lg",
          success: "!border-emerald-200 !bg-emerald-50 !text-emerald-950 dark:!border-emerald-800 dark:!bg-emerald-950 dark:!text-emerald-100",
          info: "!border-blue-200 !bg-blue-50 !text-blue-950 dark:!border-blue-800 dark:!bg-blue-950 dark:!text-blue-100",
          warning: "!border-amber-200 !bg-amber-50 !text-amber-950 dark:!border-amber-800 dark:!bg-amber-950 dark:!text-amber-100",
          error: "!border-red-200 !bg-red-50 !text-red-950 dark:!border-red-800 dark:!bg-red-950 dark:!text-red-100",
          actionButton: "!bg-foreground !text-background",
          cancelButton: "!bg-muted !text-muted-foreground",
        },
      }}
      {...props}
    />
  )
}

export { Toaster }
