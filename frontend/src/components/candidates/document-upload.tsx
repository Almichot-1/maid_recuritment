"use client"

import * as React from "react"
import { FileRejection, useDropzone } from "react-dropzone"
import { UploadCloud, File as FileIcon, X, CheckCircle2, AlertCircle } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

export interface DocumentUploadProps {
  candidateId?: string // Optional because we might be in "Create" mode
  documentType: string
  accept: Record<string, string[]>
  maxSize: number // in bytes
  onUpload?: (file: File) => void | Promise<unknown> // Triggered when file naturally resolves
  onRemove?: () => void
  title: string
  description?: string
  mode?: "deferred" | "instant"
  disabled?: boolean
}

export function DocumentUpload({
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
  const [status, setStatus] = React.useState<"empty" | "selected" | "error">("empty")
  const [errorMessage, setErrorMessage] = React.useState<string | null>(null)

  const onDrop = React.useCallback(
    async (acceptedFiles: File[], fileRejections: FileRejection[]) => {
      if (fileRejections.length > 0) {
        setStatus("error")
        setErrorMessage(fileRejections[0].errors[0]?.message || "File rejected")
        return
      }

      if (acceptedFiles.length > 0) {
        const selectedFile = acceptedFiles[0]
        let objectUrl: string | null = null
        setFile(selectedFile)
        setStatus("selected")
        setErrorMessage(null)

        // Create preview
        if (selectedFile.type.startsWith("image/") || selectedFile.type.startsWith("video/")) {
          objectUrl = URL.createObjectURL(selectedFile)
          setPreview(objectUrl)
        }

        if (onUpload) {
          try {
            await onUpload(selectedFile)
          } catch (error) {
            if (objectUrl) {
              URL.revokeObjectURL(objectUrl)
            }
            setPreview(null)
            setFile(null)
            setStatus("error")
            setErrorMessage(error instanceof Error ? error.message : "Upload failed")
          }
        }
      }
    },
    [onUpload]
  )

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept,
    maxSize,
    maxFiles: 1,
    disabled: disabled || status === "selected",
  })

  const handleRemove = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (preview) URL.revokeObjectURL(preview)
    setFile(null)
    setPreview(null)
    setStatus("empty")
    setErrorMessage(null)
    if (onRemove) onRemove()
  }

  React.useEffect(() => {
    // Cleanup memory on unmount
    return () => {
      if (preview) URL.revokeObjectURL(preview)
    }
  }, [preview])

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
          {title}
        </label>
        {status === "selected" && (
          <span className="flex items-center text-xs font-medium text-green-600 dark:text-green-400">
            <span className="mr-2 h-2 w-2 rounded-full bg-green-500 animate-pulse" />
            <CheckCircle2 className="mr-1 h-3.5 w-3.5" />
            {mode === "instant" ? "Uploading" : "Queued"}
          </span>
        )}
      </div>

      <div
        {...getRootProps()}
        className={cn(
          "relative flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed p-6 transition-all duration-200 motion-safe:hover:-translate-y-0.5",
          isDragActive ? "border-primary bg-primary/5" : "border-muted-foreground/25 hover:bg-muted/50 hover:border-muted-foreground/50",
          status === "selected" ? "border-green-500/50 bg-green-50/50 dark:bg-green-950/10" : "",
          status === "error" ? "border-destructive/50 bg-destructive/5 dark:bg-destructive/10" : "",
          status === "selected" || disabled ? "cursor-default opacity-80" : ""
        )}
      >
        <input {...getInputProps()} />

        {status === "empty" && (
          <div className="flex flex-col items-center justify-center text-center space-y-2">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
              <UploadCloud className="h-6 w-6 text-primary" />
            </div>
            <div className="text-sm">
              <span className="font-semibold text-primary">Click to upload</span> or drag and drop
            </div>
            {description && <p className="text-xs text-muted-foreground">{description}</p>}
          </div>
        )}

        {status === "selected" && file && (
          <div className="flex w-full flex-col items-center space-y-4">
            {preview ? (
              <div className="relative w-full overflow-hidden rounded-md bg-black/5 flex justify-center p-2 border border-black/10 dark:border-white/10 shadow-sm">
                {file.type.startsWith("video/") ? (
                  <video src={preview} controls className="h-32 rounded object-cover" />
                ) : file.type.startsWith("image/") ? (
                  <img src={preview} alt="Preview" className="h-32 rounded object-cover shadow-sm" />
                ) : (
                  <div className="flex h-32 w-full items-center justify-center bg-muted rounded">
                    <FileIcon className="h-10 w-10 text-muted-foreground" />
                  </div>
                )}
              </div>
            ) : (
              <div className="flex items-center justify-center w-full h-16 bg-muted rounded">
                <FileIcon className="h-6 w-6 text-muted-foreground" />
              </div>
            )}
            <div className="flex w-full items-center justify-between rounded-md bg-background px-3 py-2 border shadow-sm">
              <div className="flex items-center truncate">
                <FileIcon className="mr-2 h-4 w-4 text-blue-500 shrink-0" />
                <span className="text-xs font-medium truncate">{file.name}</span>
                <span className="text-[10px] text-muted-foreground ml-2">({(file.size / (1024 * 1024)).toFixed(2)} MB)</span>
              </div>
              <Button variant="ghost" size="icon" onClick={handleRemove} className="h-7 w-7 text-muted-foreground hover:text-destructive pointer-events-auto shrink-0">
                <X className="h-4 w-4" />
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              {mode === "instant"
                ? "This file is uploading now."
                : "This file is queued and will upload right after the candidate record is saved."}
            </p>
          </div>
        )}

        {status === "error" && (
          <div className="flex flex-col items-center justify-center text-center space-y-2 pointer-events-auto">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircle className="h-5 w-5 text-destructive" />
            </div>
            <p className="text-sm font-semibold text-destructive">Upload failed</p>
            <p className="text-xs text-muted-foreground max-w-[250px] truncate">{errorMessage}</p>
            <Button variant="outline" size="sm" onClick={(e) => { e.stopPropagation(); setStatus("empty") }} className="mt-2 h-8">
              Try Again
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
