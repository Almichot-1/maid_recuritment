"use client"

import * as React from "react"
import { UploadCloud, File as FileIcon, X, CheckCircle2, AlertCircle, Loader2 } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

export interface DocumentUploadProps {
  candidateId?: string
  documentType: string
  accept: Record<string, string[]>
  maxSize: number
  onUpload?: (file: File) => void | Promise<unknown>
  onRemove?: () => void
  title: string
  description?: string
  mode?: "deferred" | "instant"
  disabled?: boolean
}

export function DocumentUpload({
  documentType,
  accept,
  maxSize,
  onUpload,
  onRemove,
  title,
  description,
  mode = "deferred",
  disabled = false,
}: DocumentUploadProps) {
  const [file, setFile] = React.useState<File | null>(null)
  const [preview, setPreview] = React.useState<string | null>(null)
  const [status, setStatus] = React.useState<"empty" | "selected" | "uploading" | "error">("empty")
  const [errorMessage, setErrorMessage] = React.useState<string | null>(null)
  const [isDragActive, setIsDragActive] = React.useState(false)
  const inputRef = React.useRef<HTMLInputElement | null>(null)
  const shouldRenderPreview = React.useCallback(
    (selectedFile: File) => {
      if (selectedFile.type.startsWith("video/")) {
        return true
      }

      return documentType === "photo" && selectedFile.type.startsWith("image/")
    },
    [documentType]
  )

  const acceptValue = React.useMemo(
    () => {
      const extensions = Object.values(accept).flat().filter(Boolean)
      if (extensions.length > 0) {
        return Array.from(new Set(extensions.map((extension) => extension.toLowerCase()))).join(",")
      }
      return Object.keys(accept).join(",")
    },
    [accept]
  )

  const isAcceptedFile = React.useCallback(
    (selectedFile: File) => {
      const normalizedType = selectedFile.type.toLowerCase()
      const extension = (() => {
        const normalizedName = selectedFile.name.toLowerCase()
        const separatorIndex = normalizedName.lastIndexOf(".")
        return separatorIndex >= 0 ? normalizedName.slice(separatorIndex) : ""
      })()

      return Object.entries(accept).some(([mimeType, extensions]) => {
        const normalizedMime = mimeType.toLowerCase()
        const mimeMatch = normalizedMime.endsWith("/*")
          ? normalizedType.startsWith(normalizedMime.slice(0, -1))
          : normalizedType === normalizedMime
        const extensionMatch = extensions.some((allowedExtension) => extension === allowedExtension.toLowerCase())
        return mimeMatch || extensionMatch
      })
    },
    [accept]
  )

  const clearSelectedFile = React.useCallback(() => {
    if (preview) {
      URL.revokeObjectURL(preview)
    }
    setFile(null)
    setPreview(null)
    setStatus("empty")
    setErrorMessage(null)
    if (inputRef.current) {
      inputRef.current.value = ""
    }
  }, [preview])

  const processFile = React.useCallback(
    async (selectedFile: File | null) => {
      if (!selectedFile) {
        return
      }

      if (!isAcceptedFile(selectedFile)) {
        setStatus("error")
        setErrorMessage("This file type is not allowed here.")
        return
      }

      if (selectedFile.size > maxSize) {
        setStatus("error")
        setErrorMessage(`File is too large. Maximum size is ${(maxSize / (1024 * 1024)).toFixed(0)} MB.`)
        return
      }

      let objectURL: string | null = null
      setFile(selectedFile)
      setStatus(mode === "instant" ? "uploading" : "selected")
      setErrorMessage(null)

      if (shouldRenderPreview(selectedFile)) {
        objectURL = URL.createObjectURL(selectedFile)
        setPreview(objectURL)
      } else {
        setPreview(null)
      }

      if (!onUpload) {
        setStatus("selected")
        return
      }

      try {
        await onUpload(selectedFile)
        setStatus("selected")
      } catch (error) {
        if (objectURL) {
          URL.revokeObjectURL(objectURL)
        }
        setPreview(null)
        setFile(null)
        setStatus("error")
        setErrorMessage(error instanceof Error ? error.message : "Upload failed")
      } finally {
        if (inputRef.current) {
          inputRef.current.value = ""
        }
      }
    },
    [isAcceptedFile, maxSize, mode, onUpload, shouldRenderPreview]
  )

  const scheduleProcessing = React.useCallback((selectedFile: File | null) => {
    window.setTimeout(() => {
      void processFile(selectedFile)
    }, 0)
  }, [processFile])

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = event.target.files?.[0] || null
    scheduleProcessing(selectedFile)
  }

  const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    if (disabled || status === "selected" || status === "uploading") {
      return
    }
    setIsDragActive(false)
    const selectedFile = event.dataTransfer.files?.[0] || null
    scheduleProcessing(selectedFile)
  }

  const handleRemove = (event: React.MouseEvent) => {
    event.stopPropagation()
    clearSelectedFile()
    onRemove?.()
  }

  React.useEffect(() => {
    return () => {
      if (preview) {
        URL.revokeObjectURL(preview)
      }
    }
  }, [preview])

  const pickerClassName =
    "w-full max-w-xs rounded-md border border-border/70 bg-background px-3 py-2 text-sm text-foreground file:mr-3 file:rounded-md file:border-0 file:bg-primary file:px-3 file:py-2 file:text-sm file:font-medium file:text-primary-foreground hover:file:opacity-90 focus:outline-none focus:ring-2 focus:ring-primary/30"

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
          {title}
        </label>
        {status === "selected" && (
          <span className="flex items-center text-xs font-medium text-green-600 dark:text-green-400">
            <span className="mr-2 h-2 w-2 animate-pulse rounded-full bg-green-500" />
            <CheckCircle2 className="mr-1 h-3.5 w-3.5" />
            {mode === "instant" ? "Uploaded" : "Queued"}
          </span>
        )}
        {status === "uploading" && (
          <span className="flex items-center text-xs font-medium text-primary">
            <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" />
            Uploading
          </span>
        )}
      </div>

      <div
        onDragOver={(event) => {
          event.preventDefault()
          if (!disabled && status !== "selected" && status !== "uploading") {
            setIsDragActive(true)
          }
        }}
        onDragLeave={(event) => {
          event.preventDefault()
          setIsDragActive(false)
        }}
        onDrop={handleDrop}
        className={cn(
          "relative flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-6 transition-all duration-200 motion-safe:hover:-translate-y-0.5",
          isDragActive ? "border-primary bg-primary/5" : "border-muted-foreground/25 hover:border-muted-foreground/50 hover:bg-muted/50",
          status === "selected" ? "border-green-500/50 bg-green-50/50 dark:bg-green-950/10" : "",
          status === "error" ? "border-destructive/50 bg-destructive/5 dark:bg-destructive/10" : "",
          status === "selected" || status === "uploading" || disabled ? "cursor-default opacity-80" : "cursor-pointer"
        )}
      >
        {status === "empty" && (
          <div className="flex flex-col items-center justify-center space-y-2 text-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
              <UploadCloud className="h-6 w-6 text-primary" />
            </div>
            <div className="text-sm">
              <span className="font-semibold text-primary">Drag and drop here</span> or use the file picker
            </div>
            {description ? <p className="text-xs text-muted-foreground">{description}</p> : null}
            <input
              ref={inputRef}
              type="file"
              accept={acceptValue}
              onChange={handleInputChange}
              disabled={disabled}
              className={pickerClassName}
              onClick={(event) => event.stopPropagation()}
            />
          </div>
        )}

        {(status === "selected" || status === "uploading") && file ? (
          <div className="flex w-full flex-col items-center space-y-4">
            {preview ? (
              <div className="relative flex w-full justify-center overflow-hidden rounded-md border border-black/10 bg-black/5 p-2 shadow-sm dark:border-white/10">
                {file.type.startsWith("video/") ? (
                  <video src={preview} controls className="h-32 rounded object-cover" />
                ) : file.type.startsWith("image/") ? (
                  <img src={preview} alt="Preview" className="h-32 rounded object-cover shadow-sm" />
                ) : (
                  <div className="flex h-32 w-full items-center justify-center rounded bg-muted">
                    <FileIcon className="h-10 w-10 text-muted-foreground" />
                  </div>
                )}
              </div>
            ) : (
              <div className="flex h-16 w-full items-center justify-center rounded bg-muted">
                <FileIcon className="h-6 w-6 text-muted-foreground" />
              </div>
            )}
            <div className="flex w-full items-center justify-between rounded-md border bg-background px-3 py-2 shadow-sm">
              <div className="flex items-center truncate">
                <FileIcon className="mr-2 h-4 w-4 shrink-0 text-blue-500" />
                <span className="truncate text-xs font-medium">{file.name}</span>
                <span className="ml-2 text-[10px] text-muted-foreground">({(file.size / (1024 * 1024)).toFixed(2)} MB)</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleRemove}
                disabled={status === "uploading"}
                className="h-7 w-7 shrink-0 pointer-events-auto text-muted-foreground hover:text-destructive"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              {status === "uploading"
                ? "This file is uploading now."
                : mode === "instant"
                  ? "This file has been uploaded successfully."
                  : "This file is queued and will upload right after the candidate record is saved."}
            </p>
          </div>
        ) : null}

        {status === "error" && (
          <div className="pointer-events-auto flex flex-col items-center justify-center space-y-2 text-center">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircle className="h-5 w-5 text-destructive" />
            </div>
            <p className="text-sm font-semibold text-destructive">Upload failed</p>
            <p className="max-w-[250px] truncate text-xs text-muted-foreground">{errorMessage}</p>
            <div className="flex w-full max-w-xs flex-col items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={(event) => {
                  event.stopPropagation()
                  clearSelectedFile()
                }}
                className="mt-2 h-8"
              >
                Reset
              </Button>
              <input
                ref={inputRef}
                type="file"
                accept={acceptValue}
                onChange={handleInputChange}
                disabled={disabled}
                className={pickerClassName}
                onClick={(event) => event.stopPropagation()}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
