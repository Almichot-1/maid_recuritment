"use client";

import * as React from "react";
import { AxiosError } from "axios";
import { format, formatDistanceToNow } from "date-fns";
import { useRouter, useSearchParams } from "next/navigation";
import { Loader2, MessageSquare, RefreshCw, Send, Wifi, WifiOff } from "lucide-react";
import { toast } from "sonner";

import { PageHeader } from "@/components/layout/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { useCurrentUser } from "@/hooks/use-auth";
import {
  isChatStorageUnavailableError,
  useCandidateChatThread,
  useChatMessages,
  useChatSummary,
  useChatRealtime,
  useChatThreads,
  useMarkChatThreadRead,
  useSendChatMessage,
  useWorkspaceChatThread,
} from "@/hooks/use-chat";
import { usePairingContext } from "@/hooks/use-pairings";
import { cn } from "@/lib/utils";
import { ChatMessage, ChatThreadSummary } from "@/types";

function selectedThreadStorageKey(pairingId: string) {
  return `chat-selected-thread:${pairingId}`;
}

function draftStorageKey(threadId: string) {
  return `chat-draft:${threadId}`;
}

function readFromLocalStorage(key: string) {
  if (typeof window === "undefined") {
    return "";
  }

  return window.localStorage.getItem(key) || "";
}

function writeToLocalStorage(key: string, value: string) {
  if (typeof window === "undefined") {
    return;
  }

  if (!value) {
    window.localStorage.removeItem(key);
    return;
  }

  window.localStorage.setItem(key, value);
}

function getThreadTitle(thread: ChatThreadSummary) {
  if (thread.scope_type === "workspace") {
    return "Workspace chat";
  }

  return thread.candidate_name || "Candidate chat";
}

function getThreadContext(thread: ChatThreadSummary, fallbackName = "Partner workspace") {
  if (thread.scope_type === "workspace") {
    return thread.partner_agency?.company_name || thread.partner_agency?.full_name || fallbackName;
  }

  return `With ${thread.partner_agency?.company_name || thread.partner_agency?.full_name || fallbackName}`;
}

function groupMessagesByDay(messages: ChatMessage[]) {
  const groups: Array<{ key: string; label: string; messages: ChatMessage[] }> = [];

  messages.forEach((message) => {
    const date = new Date(message.created_at);
    const key = format(date, "yyyy-MM-dd");
    const lastGroup = groups[groups.length - 1];

    if (lastGroup && lastGroup.key === key) {
      lastGroup.messages.push(message);
      return;
    }

    groups.push({
      key,
      label: format(date, "MMM dd, yyyy"),
      messages: [message],
    });
  });

  return groups;
}

export default function PartnersChatPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const candidateIdFromQuery = searchParams.get("candidate_id") || undefined;
  const { user } = useCurrentUser();
  const { hasActivePairs, isReady, activePairingId, activeWorkspace } = usePairingContext();

  const {
    data: threads = [],
    isLoading: isThreadsLoading,
    isError: isThreadsError,
    error: threadsError,
    refetch: refetchThreads,
  } = useChatThreads();
  const { data: workspaceThreadFallback, error: workspaceThreadError } = useWorkspaceChatThread();
  const candidateThreadQuery = useCandidateChatThread(candidateIdFromQuery, {
    enabled: Boolean(candidateIdFromQuery),
  });
  const chatSummaryQuery = useChatSummary();
  const realtime = useChatRealtime();

  const workspaceThread = React.useMemo(
    () => threads.find((thread) => thread.scope_type === "workspace") || workspaceThreadFallback || null,
    [threads, workspaceThreadFallback]
  );

  const orderedThreads = React.useMemo(() => {
    if (!threads.length) {
      return workspaceThread ? [workspaceThread] : [];
    }

    const workspace = threads.find((thread) => thread.scope_type === "workspace") || workspaceThread;
    const candidates = threads.filter((thread) => thread.scope_type === "candidate");

    if (!workspace) {
      return candidates;
    }

    return [workspace, ...candidates.filter((thread) => thread.id !== workspace.id)];
  }, [threads, workspaceThread]);

  const [selectedThreadId, setSelectedThreadId] = React.useState<string | null>(null);
  const [composerValue, setComposerValue] = React.useState("");
  const previousPairingIdRef = React.useRef<string | null>(null);
  const pendingWorkspaceResetPairingRef = React.useRef<string | null>(null);
  const candidateParamStateRef = React.useRef<string | null>(null);
  const messagesContainerRef = React.useRef<HTMLDivElement | null>(null);
  const readReceiptMarkerRef = React.useRef<string | null>(null);

  const activeThread = React.useMemo(
    () => orderedThreads.find((thread) => thread.id === selectedThreadId) || null,
    [orderedThreads, selectedThreadId]
  );

  const messagesQuery = useChatMessages(activeThread?.id);
  const sendMessage = useSendChatMessage(activeThread?.id);
  const markThreadRead = useMarkChatThreadRead(activeThread?.id);

  const chatStorageUnavailable = React.useMemo(
    () =>
      [threadsError, workspaceThreadError, candidateThreadQuery.error, chatSummaryQuery.error, messagesQuery.error].some(
        (error) => isChatStorageUnavailableError(error)
      ),
    [threadsError, workspaceThreadError, candidateThreadQuery.error, chatSummaryQuery.error, messagesQuery.error]
  );

  const messageGroups = React.useMemo(
    () => groupMessagesByDay(messagesQuery.messages),
    [messagesQuery.messages]
  );

  React.useEffect(() => {
    if (!activePairingId) {
      setSelectedThreadId(null);
      return;
    }

    const pairingChanged =
      previousPairingIdRef.current !== null && previousPairingIdRef.current !== activePairingId;

    if (pairingChanged) {
      pendingWorkspaceResetPairingRef.current = activePairingId;
      const nextThreadId = workspaceThread?.id || null;
      setSelectedThreadId(nextThreadId);

      if (nextThreadId) {
        writeToLocalStorage(selectedThreadStorageKey(activePairingId), nextThreadId);
        pendingWorkspaceResetPairingRef.current = null;
      }

      previousPairingIdRef.current = activePairingId;
      return;
    }

    if (!selectedThreadId) {
      if (pendingWorkspaceResetPairingRef.current === activePairingId) {
        if (workspaceThread?.id) {
          setSelectedThreadId(workspaceThread.id);
          writeToLocalStorage(selectedThreadStorageKey(activePairingId), workspaceThread.id);
          pendingWorkspaceResetPairingRef.current = null;
        }
        previousPairingIdRef.current = activePairingId;
        return;
      }

      const storedThreadId = readFromLocalStorage(selectedThreadStorageKey(activePairingId));
      const exists = orderedThreads.some((thread) => thread.id === storedThreadId);
      const nextThreadId = exists ? storedThreadId : workspaceThread?.id || orderedThreads[0]?.id || null;

      if (nextThreadId) {
        setSelectedThreadId(nextThreadId);
        writeToLocalStorage(selectedThreadStorageKey(activePairingId), nextThreadId);
      }
    }

    previousPairingIdRef.current = activePairingId;
  }, [activePairingId, orderedThreads, selectedThreadId, workspaceThread?.id]);

  React.useEffect(() => {
    if (!activePairingId || !selectedThreadId) {
      return;
    }

    writeToLocalStorage(selectedThreadStorageKey(activePairingId), selectedThreadId);
  }, [activePairingId, selectedThreadId]);

  React.useEffect(() => {
    if (!candidateIdFromQuery) {
      candidateParamStateRef.current = null;
      return;
    }

    if (candidateThreadQuery.isSuccess && candidateThreadQuery.data?.id) {
      setSelectedThreadId(candidateThreadQuery.data.id);
      candidateParamStateRef.current = `ok:${candidateIdFromQuery}`;
      return;
    }

    if (candidateThreadQuery.isError && candidateParamStateRef.current !== `err:${candidateIdFromQuery}`) {
      candidateParamStateRef.current = `err:${candidateIdFromQuery}`;

      if (isChatStorageUnavailableError(candidateThreadQuery.error)) {
        toast.error("Chat is unavailable in this environment right now.");
        router.replace("/partners/chat");
        return;
      }

      const status = (candidateThreadQuery.error as AxiosError)?.response?.status;
      if (status === 403 || status === 404) {
        toast.error("This candidate chat is not available in the current workspace.");
      } else {
        toast.error("Unable to open this candidate chat right now.");
      }

      router.replace("/partners/chat");
    }
  }, [candidateIdFromQuery, candidateThreadQuery.data?.id, candidateThreadQuery.error, candidateThreadQuery.isError, candidateThreadQuery.isSuccess, router]);

  React.useEffect(() => {
    if (!activeThread?.id) {
      setComposerValue("");
      return;
    }

    setComposerValue(readFromLocalStorage(draftStorageKey(activeThread.id)));
  }, [activeThread?.id]);

  React.useEffect(() => {
    const container = messagesContainerRef.current;
    if (!container) {
      return;
    }

    requestAnimationFrame(() => {
      container.scrollTop = container.scrollHeight;
    });
  }, [activeThread?.id]);

  React.useEffect(() => {
    if (!activeThread?.id || !messagesQuery.isSuccess || activeThread.unread_count <= 0) {
      return;
    }

    const marker = `${activeThread.id}:${messagesQuery.messages.length}`;
    if (readReceiptMarkerRef.current === marker) {
      return;
    }

    readReceiptMarkerRef.current = marker;
    markThreadRead.mutate();
  }, [activeThread?.id, activeThread?.unread_count, markThreadRead, messagesQuery.isSuccess, messagesQuery.messages.length]);

  const handleSelectThread = React.useCallback(
    (threadId: string) => {
      setSelectedThreadId(threadId);
    },
    [setSelectedThreadId]
  );

  const handleComposerChange = React.useCallback(
    (value: string) => {
      setComposerValue(value);
      if (activeThread?.id) {
        writeToLocalStorage(draftStorageKey(activeThread.id), value);
      }
    },
    [activeThread?.id]
  );

  const handleSendMessage = React.useCallback(async () => {
    if (!activeThread?.id || sendMessage.isPending) {
      return;
    }

    const trimmed = composerValue.trim();
    if (!trimmed) {
      return;
    }

    try {
      await sendMessage.mutateAsync(trimmed);
      setComposerValue("");
      writeToLocalStorage(draftStorageKey(activeThread.id), "");

      const container = messagesContainerRef.current;
      if (container) {
        requestAnimationFrame(() => {
          container.scrollTop = container.scrollHeight;
        });
      }
    } catch {
      // Mutations are already surfaced via toast.
    }
  }, [activeThread?.id, composerValue, sendMessage]);

  const handleLoadOlder = React.useCallback(async () => {
    if (!messagesQuery.hasOlder || messagesQuery.isFetchingNextPage) {
      return;
    }

    const container = messagesContainerRef.current;
    const previousHeight = container?.scrollHeight || 0;
    const previousTop = container?.scrollTop || 0;

    await messagesQuery.loadOlder();

    if (!container) {
      return;
    }

    requestAnimationFrame(() => {
      const currentHeight = container.scrollHeight;
      container.scrollTop = previousTop + (currentHeight - previousHeight);
    });
  }, [messagesQuery]);

  if (!isReady) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (!hasActivePairs || !activePairingId) {
    return (
      <div className="space-y-6">
        <PageHeader
          heading="Chat"
          text="Select or activate a partner workspace to start agency-to-agency chat."
        />
        <Card>
          <CardContent className="py-10 text-center">
            <p className="text-sm text-muted-foreground">
              No active workspace is currently selected.
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (chatStorageUnavailable) {
    return (
      <div className="space-y-6">
        <PageHeader
          heading="Chat"
          text="Workspace and candidate conversations for the active partner workspace."
        />
        <Card>
          <CardContent className="space-y-4 py-10 text-center">
            <p className="text-base font-semibold text-foreground">Chat storage is not configured for this environment.</p>
            <p className="text-sm text-muted-foreground">
              Please run the chat database migration and restart the backend, then refresh this page.
            </p>
            <div className="flex justify-center">
              <Button variant="outline" onClick={() => void Promise.all([refetchThreads(), chatSummaryQuery.refetch()])}>
                <RefreshCw className="mr-2 h-4 w-4" />
                Retry
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        heading="Chat"
        text="Workspace and candidate conversations for the active partner workspace."
        action={
          <Badge variant={realtime.isConnected ? "default" : "secondary"}>
            {realtime.isConnected ? (
              <Wifi className="mr-1 h-3.5 w-3.5" />
            ) : (
              <WifiOff className="mr-1 h-3.5 w-3.5" />
            )}
            {realtime.isConnected
              ? "Live"
              : realtime.isReconnecting
                ? "Reconnecting"
                : "Offline"}
          </Badge>
        }
      />

      {isThreadsError ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center gap-3 py-10 text-center">
            <p className="text-sm text-muted-foreground">Unable to load chat threads right now.</p>
            <Button variant="outline" onClick={() => refetchThreads()}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Retry
            </Button>
          </CardContent>
        </Card>
      ) : null}

      <div className="grid gap-4 lg:grid-cols-[320px_minmax(0,1fr)]">
        <Card className="min-w-0">
          <CardHeader className="pb-3">
            <CardTitle>Threads</CardTitle>
            <CardDescription>
              Workspace chat first, then candidate-specific conversations.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {isThreadsLoading ? (
              <div className="space-y-2">
                {Array.from({ length: 6 }).map((_, index) => (
                  <Skeleton key={index} className="h-20 w-full rounded-xl" />
                ))}
              </div>
            ) : orderedThreads.length === 0 ? (
              <div className="rounded-xl border border-dashed p-5 text-sm text-muted-foreground">
                No threads available yet for this workspace.
              </div>
            ) : (
              orderedThreads.map((thread) => {
                const isActive = thread.id === activeThread?.id;
                const timestamp = thread.last_message_at || thread.updated_at;

                return (
                  <button
                    key={thread.id}
                    type="button"
                    onClick={() => handleSelectThread(thread.id)}
                    className={cn(
                      "w-full rounded-xl border p-3 text-left transition-colors",
                      isActive
                        ? "border-primary/60 bg-primary/10"
                        : "border-border hover:bg-muted/40"
                    )}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-semibold text-foreground">
                          {getThreadTitle(thread)}
                        </p>
                        <p className="truncate text-xs text-muted-foreground">
                          {getThreadContext(thread, activeWorkspace?.partner_agency.company_name)}
                        </p>
                      </div>
                      {thread.unread_count > 0 ? (
                        <Badge className="shrink-0">{thread.unread_count}</Badge>
                      ) : null}
                    </div>
                    <div className="mt-2 space-y-1">
                      <p className="line-clamp-1 text-xs text-muted-foreground">
                        {thread.last_message_preview || "No messages yet"}
                      </p>
                      <p className="text-[11px] text-muted-foreground">
                        {formatDistanceToNow(new Date(timestamp), { addSuffix: true })}
                      </p>
                    </div>
                  </button>
                );
              })
            )}
          </CardContent>
        </Card>

        <Card className="min-w-0">
          {!activeThread ? (
            <CardContent className="flex min-h-[55vh] flex-col items-center justify-center gap-3 text-center">
              <MessageSquare className="h-8 w-8 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Select a thread to start chatting.</p>
            </CardContent>
          ) : (
            <div className="flex h-[68vh] min-h-[460px] flex-col">
              <CardHeader className="border-b pb-4">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <CardTitle className="truncate text-lg">
                      {activeThread.scope_type === "workspace"
                        ? activeThread.partner_agency?.company_name ||
                          activeThread.partner_agency?.full_name ||
                          activeWorkspace?.partner_agency.company_name ||
                          "Workspace chat"
                        : activeThread.candidate_name || "Candidate chat"}
                    </CardTitle>
                    <CardDescription className="truncate">
                      {activeThread.scope_type === "workspace"
                        ? "General conversation in this partner workspace"
                        : `Candidate thread in ${
                            activeThread.partner_agency?.company_name ||
                            activeThread.partner_agency?.full_name ||
                            activeWorkspace?.partner_agency.company_name ||
                            "current workspace"
                          }`}
                    </CardDescription>
                  </div>
                  {activeThread.unread_count > 0 ? (
                    <Badge variant="secondary">{activeThread.unread_count} unread</Badge>
                  ) : null}
                </div>
                {!realtime.isConnected ? (
                  <p className="pt-2 text-xs text-muted-foreground">
                    Live connection is {realtime.isReconnecting ? "reconnecting" : "offline"}. Messages still sync via automatic refresh.
                  </p>
                ) : null}
              </CardHeader>

              <div ref={messagesContainerRef} className="flex-1 space-y-3 overflow-y-auto px-4 py-4">
                {messagesQuery.hasOlder ? (
                  <div className="flex justify-center">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={handleLoadOlder}
                      disabled={messagesQuery.isFetchingNextPage}
                    >
                      {messagesQuery.isFetchingNextPage ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : null}
                      Load older messages
                    </Button>
                  </div>
                ) : null}

                {messagesQuery.isLoading ? (
                  <div className="space-y-3">
                    {Array.from({ length: 6 }).map((_, index) => (
                      <Skeleton key={index} className="h-16 w-full rounded-xl" />
                    ))}
                  </div>
                ) : messageGroups.length === 0 ? (
                  <div className="flex h-full items-center justify-center py-8 text-center text-sm text-muted-foreground">
                    No messages yet. Start the conversation.
                  </div>
                ) : (
                  messageGroups.map((group) => (
                    <div key={group.key} className="space-y-3">
                      <div className="flex items-center justify-center">
                        <span className="rounded-full border bg-muted px-3 py-1 text-[11px] text-muted-foreground">
                          {group.label}
                        </span>
                      </div>

                      {group.messages.map((message) => {
                        const isMine = Boolean(message.is_mine || message.sender_user_id === user?.id);

                        return (
                          <div
                            key={message.id}
                            className={cn("flex w-full", isMine ? "justify-end" : "justify-start")}
                          >
                            <div
                              className={cn(
                                "max-w-[88%] rounded-2xl px-3 py-2 text-sm sm:max-w-[78%]",
                                isMine
                                  ? "bg-primary text-primary-foreground"
                                  : "border border-border bg-muted/40"
                              )}
                            >
                              {!isMine ? (
                                <p className="mb-1 text-[11px] font-medium text-muted-foreground">
                                  {message.sender_name}
                                </p>
                              ) : null}
                              <p className="whitespace-pre-wrap break-words">{message.body}</p>
                              <p
                                className={cn(
                                  "mt-1 text-[11px]",
                                  isMine ? "text-primary-foreground/80" : "text-muted-foreground"
                                )}
                              >
                                {formatDistanceToNow(new Date(message.created_at), { addSuffix: true })}
                              </p>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  ))
                )}
              </div>

              <div className="border-t px-4 py-3">
                <div className="space-y-3">
                  <Textarea
                    value={composerValue}
                    onChange={(event) => handleComposerChange(event.target.value)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" && !event.shiftKey) {
                        event.preventDefault();
                        void handleSendMessage();
                      }
                    }}
                    placeholder="Write a message..."
                    rows={3}
                    className="max-h-40 min-h-[84px] resize-y"
                  />
                  <div className="flex items-center justify-between gap-3">
                    <p className="text-xs text-muted-foreground">Enter to send, Shift+Enter for newline</p>
                    <Button
                      type="button"
                      onClick={() => void handleSendMessage()}
                      disabled={sendMessage.isPending || composerValue.trim().length === 0}
                    >
                      {sendMessage.isPending ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : (
                        <Send className="mr-2 h-4 w-4" />
                      )}
                      Send
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
