import * as React from "react";
import { AxiosError } from "axios";
import {
  InfiniteData,
  QueryClient,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";

import api from "@/lib/api";
import { getApiBaseUrl } from "@/lib/api-base-url";
import { useCurrentUser } from "@/hooks/use-auth";
import { usePairingStore } from "@/stores/pairing-store";
import {
  ChatMessage,
  ChatReadEvent,
  ChatSocketEvent,
  ChatSummary,
  ChatThreadScope,
  ChatThreadSummary,
  PairingAgencySummary,
  UserRole,
} from "@/types";

const CHAT_PAGE_SIZE = 30;

const chatKeys = {
  all: ["chat"] as const,
  threads: (pairingId: string | null) => ["chat", "threads", pairingId] as const,
  workspaceThread: (pairingId: string | null) => ["chat", "thread", "workspace", pairingId] as const,
  candidateThread: (pairingId: string | null, candidateId?: string) =>
    ["chat", "thread", "candidate", pairingId, candidateId] as const,
  messages: (pairingId: string | null, threadId?: string) => ["chat", "messages", pairingId, threadId] as const,
  summary: (pairingId: string | null) => ["chat", "summary", pairingId] as const,
};

type ChatMessagesPage = {
  messages: ChatMessage[];
  nextBefore: string | null;
  hasMore: boolean;
};

interface ChatSummaryResponse {
  summary?: unknown;
  total_unread?: number;
  unread_count?: number;
}

interface SendMessageContext {
  previousMessages?: InfiniteData<ChatMessagesPage>;
  optimisticMessageId: string;
  previousThreads?: ChatThreadSummary[];
}

interface MarkReadContext {
  previousThreads?: ChatThreadSummary[];
  previousSummary?: ChatSummary;
}

interface ChatApiErrorPayload {
  error?: string;
  message?: string;
}

function getChatErrorStatus(error: unknown): number | undefined {
  return (error as AxiosError<ChatApiErrorPayload>)?.response?.status;
}

export function isChatStorageUnavailableError(error: unknown): boolean {
  const status = getChatErrorStatus(error);
  const payload = (error as AxiosError<ChatApiErrorPayload>)?.response?.data;
  const message = String(payload?.error || payload?.message || "").toLowerCase();

  if (status === 503) {
    return message.length === 0 || message.includes("chat storage") || message.includes("storage is unavailable");
  }

  return message.includes("chat storage") || message.includes("storage is unavailable");
}

function shouldRetryChatQuery(failureCount: number, error: unknown, maxRetries: number) {
  const status = getChatErrorStatus(error);

  if (status === 401 || status === 403 || status === 404 || status === 503) {
    return false;
  }

  if (isChatStorageUnavailableError(error)) {
    return false;
  }

  return failureCount < maxRetries;
}

function toNumber(value: unknown, fallback = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function toDateString(value: unknown) {
  if (typeof value === "string" && value.trim().length > 0) {
    return value;
  }

  return new Date().toISOString();
}

function normalizePartnerAgency(raw: unknown): PairingAgencySummary | null {
  if (!raw || typeof raw !== "object") {
    return null;
  }

  const item = raw as Record<string, unknown>;
  const roleValue = String(item.role || "");
  const role =
    roleValue === UserRole.ETHIOPIAN_AGENT || roleValue === UserRole.FOREIGN_AGENT
      ? roleValue
      : UserRole.FOREIGN_AGENT;

  return {
    id: String(item.id || ""),
    full_name: String(item.full_name || item.name || "Partner"),
    company_name: String(item.company_name || ""),
    email: String(item.email || ""),
    role,
  };
}

function normalizeThread(raw: unknown): ChatThreadSummary {
  const item = (raw || {}) as Record<string, unknown>;
  const scopeRaw = String(item.scope_type || item.scope || "workspace");
  const scopeType: ChatThreadScope = scopeRaw === "candidate" ? "candidate" : "workspace";

  const candidate = (item.candidate || {}) as Record<string, unknown>;

  return {
    id: String(item.id || ""),
    scope_type: scopeType,
    pairing_id: String(item.pairing_id || item.pairingId || ""),
    candidate_id: (item.candidate_id || candidate.id ? String(item.candidate_id || candidate.id) : null),
    candidate_name: (item.candidate_name || candidate.full_name
      ? String(item.candidate_name || candidate.full_name)
      : null),
    partner_agency: normalizePartnerAgency(item.partner_agency || item.partner || item.partnerAgency),
    last_message_preview: item.last_message_preview
      ? String(item.last_message_preview)
      : item.last_message && typeof item.last_message === "object"
        ? String((item.last_message as Record<string, unknown>).body || "")
        : null,
    last_message_at: item.last_message_at
      ? String(item.last_message_at)
      : item.last_message && typeof item.last_message === "object"
        ? String((item.last_message as Record<string, unknown>).created_at || "")
        : null,
    unread_count: toNumber(item.unread_count ?? item.unread ?? 0),
    created_at: toDateString(item.created_at),
    updated_at: toDateString(item.updated_at),
  };
}

function normalizeMessage(raw: unknown, currentUserId?: string): ChatMessage {
  const item = (raw || {}) as Record<string, unknown>;
  const sender = (item.sender || {}) as Record<string, unknown>;
  const senderUserId = String(item.sender_user_id || sender.user_id || sender.id || "");

  return {
    id: String(item.id || ""),
    thread_id: String(item.thread_id || item.threadId || ""),
    sender_user_id: senderUserId,
    sender_name: String(item.sender_name || sender.full_name || "User"),
    sender_role: String(item.sender_role || sender.role || ""),
    body: String(item.body || ""),
    created_at: toDateString(item.created_at),
    is_mine: currentUserId ? senderUserId === currentUserId : undefined,
  };
}

function normalizeSummary(raw: unknown): ChatSummary {
  const item = (raw || {}) as Record<string, unknown>;
  const totalUnread = toNumber(
    item.total_unread ?? item.unread_messages ?? item.unread_count ?? item.unread ?? 0
  );

  return {
    total_unread: totalUnread,
    workspace_unread: toNumber(item.workspace_unread ?? item.unread_threads ?? 0),
    candidate_unread: toNumber(item.candidate_unread ?? 0),
    threads_with_unread: toNumber(item.threads_with_unread ?? 0),
  };
}

function sortThreads(threads: ChatThreadSummary[]) {
  return [...threads].sort((left, right) => {
    if (left.scope_type !== right.scope_type) {
      return left.scope_type === "workspace" ? -1 : 1;
    }

    const leftDate = new Date(left.last_message_at || left.updated_at || left.created_at).getTime();
    const rightDate = new Date(right.last_message_at || right.updated_at || right.created_at).getTime();

    return rightDate - leftDate;
  });
}

function extractThreadList(payload: unknown): ChatThreadSummary[] {
  const data = payload as Record<string, unknown>;
  const candidates = Array.isArray(data?.threads)
    ? data.threads
    : Array.isArray(data?.data)
      ? data.data
      : Array.isArray(payload)
        ? payload
        : [];

  return sortThreads(candidates.map((item) => normalizeThread(item)));
}

function extractThread(payload: unknown): ChatThreadSummary {
  const data = payload as Record<string, unknown>;
  return normalizeThread(data.thread || data.data || payload);
}

function extractMessagePage(payload: unknown, currentUserId?: string): ChatMessagesPage {
  const data = payload as Record<string, unknown>;
  const source = Array.isArray(data?.messages)
    ? data.messages
    : Array.isArray(data?.data)
      ? data.data
      : Array.isArray(payload)
        ? payload
        : [];

  const messages = source
    .map((message) => normalizeMessage(message, currentUserId))
    .sort((left, right) => new Date(left.created_at).getTime() - new Date(right.created_at).getTime());

  const pagination = (data.pagination || {}) as Record<string, unknown>;
  const nextBefore =
    typeof data.next_before === "string"
      ? data.next_before
      : typeof data.next_cursor === "string"
        ? data.next_cursor
      : typeof pagination.next_before === "string"
        ? pagination.next_before
        : typeof pagination.next_cursor === "string"
          ? pagination.next_cursor
        : null;

  const hasMoreRaw = data.has_more ?? pagination.has_more;
  const hasMore = typeof hasMoreRaw === "boolean" ? hasMoreRaw : Boolean(nextBefore);

  return {
    messages,
    nextBefore,
    hasMore,
  };
}

function flattenMessagePages(pages: ChatMessagesPage[]) {
  const deduped = new Map<string, ChatMessage>();

  [...pages]
    .reverse()
    .flatMap((page) => page.messages)
    .forEach((message) => {
      const existing = deduped.get(message.id);

      if (!existing || message.id.startsWith("optimistic-")) {
        deduped.set(message.id, message);
      }
    });

  return [...deduped.values()].sort(
    (left, right) => new Date(left.created_at).getTime() - new Date(right.created_at).getTime()
  );
}

function updateThreadLastMessage(threads: ChatThreadSummary[] | undefined, threadId: string, preview: string) {
  if (!threads?.length) {
    return threads;
  }

  const now = new Date().toISOString();

  const next = threads.map((thread) =>
    thread.id === threadId
      ? {
          ...thread,
          last_message_preview: preview,
          last_message_at: now,
          updated_at: now,
        }
      : thread
  );

  return sortThreads(next);
}

function upsertThread(threads: ChatThreadSummary[] | undefined, incoming: ChatThreadSummary) {
  if (!threads?.length) {
    return [incoming];
  }

  const index = threads.findIndex((thread) => thread.id === incoming.id);
  if (index === -1) {
    return sortThreads([incoming, ...threads]);
  }

  const next = [...threads];
  next[index] = { ...next[index], ...incoming };
  return sortThreads(next);
}

function applyReadState(
  threads: ChatThreadSummary[] | undefined,
  threadId: string,
  unreadCount: number
): ChatThreadSummary[] | undefined {
  if (!threads?.length) {
    return threads;
  }

  return threads.map((thread) =>
    thread.id === threadId
      ? {
          ...thread,
          unread_count: unreadCount,
        }
      : thread
  );
}

function appendMessage(
  currentData: InfiniteData<ChatMessagesPage> | undefined,
  nextMessage: ChatMessage
): InfiniteData<ChatMessagesPage> {
  if (!currentData || currentData.pages.length === 0) {
    return {
      pages: [
        {
          messages: [nextMessage],
          nextBefore: null,
          hasMore: false,
        },
      ],
      pageParams: [null],
    };
  }

  const pages = [...currentData.pages];
  const firstPage = pages[0];
  const alreadyExists = firstPage.messages.some((message) => message.id === nextMessage.id);
  if (!alreadyExists) {
    pages[0] = {
      ...firstPage,
      messages: [...firstPage.messages, nextMessage],
    };
  }

  return {
    ...currentData,
    pages,
  };
}

function replaceOptimisticMessage(
  currentData: InfiniteData<ChatMessagesPage> | undefined,
  optimisticId: string,
  serverMessage: ChatMessage
): InfiniteData<ChatMessagesPage> | undefined {
  if (!currentData) {
    return currentData;
  }

  return {
    ...currentData,
    pages: currentData.pages.map((page) => ({
      ...page,
      messages: page.messages.map((message) =>
        message.id === optimisticId ? { ...serverMessage } : message
      ),
    })),
  };
}

function buildChatWebSocketUrl(pairingId: string) {
  const apiUrl = new URL(getApiBaseUrl());
  apiUrl.protocol = apiUrl.protocol === "https:" ? "wss:" : "ws:";

  const trimmedPath = apiUrl.pathname.replace(/\/+$/, "");
  const withoutApiVersion = trimmedPath.replace(/\/api\/v\d+$/i, "");

  apiUrl.pathname = `${withoutApiVersion}/ws/chat`;
  apiUrl.search = "";
  apiUrl.searchParams.set("pairing_id", pairingId);

  return apiUrl.toString();
}

function parseSocketEvent(rawPayload: unknown, currentUserId?: string): ChatSocketEvent {
  if (typeof rawPayload !== "string") {
    return {
      type: "unknown",
      raw: rawPayload,
    };
  }

  try {
    const parsed = JSON.parse(rawPayload) as Record<string, unknown>;
    const type = String(parsed.type || parsed.event || "");
    const payload =
      parsed.payload && typeof parsed.payload === "object"
        ? (parsed.payload as Record<string, unknown>)
        : parsed;

    if (type === "connected") {
      return {
        type: "connected",
        pairing_id:
          typeof payload.pairing_id === "string"
            ? payload.pairing_id
            : typeof parsed.pairing_id === "string"
              ? parsed.pairing_id
              : undefined,
      };
    }

    if (type === "message.created") {
      const thread = payload.thread ? normalizeThread(payload.thread) : undefined;
      const message = normalizeMessage(payload.message || payload.data || payload, currentUserId);

      return {
        type: "message.created",
        pairing_id:
          typeof payload.pairing_id === "string"
            ? payload.pairing_id
            : typeof parsed.pairing_id === "string"
              ? parsed.pairing_id
              : undefined,
        thread_id: String(payload.thread_id || parsed.thread_id || thread?.id || message.thread_id || ""),
        message,
        thread,
        summary: (payload.summary || parsed.summary)
          ? normalizeSummary(payload.summary || parsed.summary)
          : undefined,
      };
    }

    if (type === "thread.read") {
      const readStateRaw = (payload.read_state || payload || {}) as Record<string, unknown>;
      const unreadRaw = readStateRaw.unread_count;
      const unreadCount =
        typeof unreadRaw === "number"
          ? unreadRaw
          : typeof unreadRaw === "string" && unreadRaw.trim().length > 0
            ? Number(unreadRaw)
            : Number.NaN;

      const readEvent: ChatReadEvent = {
        type: "thread.read",
        pairing_id:
          typeof payload.pairing_id === "string"
            ? payload.pairing_id
            : typeof parsed.pairing_id === "string"
              ? parsed.pairing_id
              : undefined,
        thread_id: String(payload.thread_id || parsed.thread_id || readStateRaw.thread_id || ""),
        read_state: {
          thread_id: String(readStateRaw.thread_id || payload.thread_id || parsed.thread_id || ""),
          unread_count: unreadCount,
          last_read_at:
            typeof readStateRaw.last_read_at === "string" ? readStateRaw.last_read_at : undefined,
        },
        summary: (payload.summary || parsed.summary)
          ? normalizeSummary(payload.summary || parsed.summary)
          : undefined,
      };

      return readEvent;
    }

    if (type === "thread.summary_updated") {
      return {
        type: "thread.summary_updated",
        pairing_id:
          typeof payload.pairing_id === "string"
            ? payload.pairing_id
            : typeof parsed.pairing_id === "string"
              ? parsed.pairing_id
              : undefined,
        thread: payload.thread ? normalizeThread(payload.thread) : undefined,
        summary: (payload.summary || parsed.summary)
          ? normalizeSummary(payload.summary || parsed.summary)
          : undefined,
      };
    }

    return {
      type: "unknown",
      raw: parsed,
    };
  } catch {
    return {
      type: "unknown",
      raw: rawPayload,
    };
  }
}

function applySocketEventToCache(
  queryClient: QueryClient,
  pairingId: string,
  event: ChatSocketEvent,
  currentUserId?: string
) {
  switch (event.type) {
    case "connected": {
      queryClient.invalidateQueries({ queryKey: chatKeys.threads(pairingId) });
      queryClient.invalidateQueries({ queryKey: chatKeys.summary(pairingId) });
      return;
    }
    case "message.created": {
      if (event.thread_id) {
        queryClient.setQueryData<InfiniteData<ChatMessagesPage>>(
          chatKeys.messages(pairingId, event.thread_id),
          (current) => appendMessage(current, normalizeMessage(event.message, currentUserId))
        );
      }

      if (event.thread) {
        queryClient.setQueryData<ChatThreadSummary[]>(chatKeys.threads(pairingId), (threads) =>
          upsertThread(threads, event.thread as ChatThreadSummary)
        );

        if (event.thread.scope_type === "workspace") {
          queryClient.setQueryData(chatKeys.workspaceThread(pairingId), event.thread);
        }

        if (event.thread.scope_type === "candidate" && event.thread.candidate_id) {
          queryClient.setQueryData(
            chatKeys.candidateThread(pairingId, event.thread.candidate_id),
            event.thread
          );
        }
      } else {
        queryClient.invalidateQueries({ queryKey: chatKeys.threads(pairingId) });
      }

      if (event.summary) {
        queryClient.setQueryData(chatKeys.summary(pairingId), event.summary);
      } else {
        queryClient.invalidateQueries({ queryKey: chatKeys.summary(pairingId) });
      }

      return;
    }
    case "thread.read": {
      const unreadCount = event.read_state?.unread_count;
      if (typeof unreadCount === "number" && Number.isFinite(unreadCount)) {
        queryClient.setQueryData<ChatThreadSummary[]>(chatKeys.threads(pairingId), (threads) =>
          applyReadState(threads, event.thread_id, unreadCount)
        );
      }

      if (event.summary) {
        queryClient.setQueryData(chatKeys.summary(pairingId), event.summary);
      } else {
        queryClient.invalidateQueries({ queryKey: chatKeys.summary(pairingId) });
      }

      return;
    }
    case "thread.summary_updated": {
      if (event.thread) {
        queryClient.setQueryData<ChatThreadSummary[]>(chatKeys.threads(pairingId), (threads) =>
          upsertThread(threads, event.thread as ChatThreadSummary)
        );
      } else {
        queryClient.invalidateQueries({ queryKey: chatKeys.threads(pairingId) });
      }

      if (event.summary) {
        queryClient.setQueryData(chatKeys.summary(pairingId), event.summary);
      } else {
        queryClient.invalidateQueries({ queryKey: chatKeys.summary(pairingId) });
      }
      return;
    }
    default:
      return;
  }
}

type ChatRealtimeStatus = "idle" | "connecting" | "connected" | "reconnecting" | "disconnected";

type SocketSubscriber = {
  onStatus: (status: ChatRealtimeStatus) => void;
  onEvent: (event: ChatSocketEvent) => void;
};

let activeSocket: WebSocket | null = null;
let activeSocketPairingId: string | null = null;
let activeSocketStatus: ChatRealtimeStatus = "idle";
let reconnectAttempt = 0;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let intentionallyClosed = false;
const socketSubscribers = new Set<SocketSubscriber>();
const maxReconnectAttempts = 6;

function publishSocketStatus(status: ChatRealtimeStatus) {
  activeSocketStatus = status;
  socketSubscribers.forEach((subscriber) => subscriber.onStatus(status));
}

function publishSocketEvent(event: ChatSocketEvent) {
  socketSubscribers.forEach((subscriber) => subscriber.onEvent(event));
}

function clearReconnectTimer() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

function teardownSocket() {
  clearReconnectTimer();
  intentionallyClosed = true;

  if (activeSocket) {
    activeSocket.close();
  }

  activeSocket = null;
  activeSocketPairingId = null;
  reconnectAttempt = 0;
  publishSocketStatus("idle");
}

function scheduleReconnect(pairingId: string) {
  if (socketSubscribers.size === 0) {
    publishSocketStatus("disconnected");
    return;
  }

  if (reconnectAttempt >= maxReconnectAttempts) {
    publishSocketStatus("disconnected");
    return;
  }

  clearReconnectTimer();
  const delays = [1000, 2000, 5000, 10000, 15000];
  const delay = delays[Math.min(reconnectAttempt, delays.length - 1)];
  reconnectAttempt += 1;

  reconnectTimer = setTimeout(() => {
    openSocket(pairingId);
  }, delay);
}

function openSocket(pairingId: string) {
  if (typeof window === "undefined") {
    return;
  }

  intentionallyClosed = false;
  publishSocketStatus(reconnectAttempt > 0 ? "reconnecting" : "connecting");

  try {
    const socket = new WebSocket(buildChatWebSocketUrl(pairingId));
    activeSocket = socket;
    activeSocketPairingId = pairingId;

    socket.onopen = () => {
      reconnectAttempt = 0;
      publishSocketStatus("connected");
      publishSocketEvent({
        type: "connected",
        pairing_id: pairingId,
      });
    };

    socket.onmessage = (event) => {
      publishSocketEvent(parseSocketEvent(event.data));
    };

    socket.onerror = () => {
      publishSocketStatus("reconnecting");
    };

    socket.onclose = () => {
      activeSocket = null;

      if (intentionallyClosed) {
        publishSocketStatus("disconnected");
        return;
      }

      publishSocketStatus("reconnecting");
      scheduleReconnect(pairingId);
    };
  } catch {
    publishSocketStatus("reconnecting");
    scheduleReconnect(pairingId);
  }
}

function ensureSocket(pairingId: string) {
  if (typeof window === "undefined") {
    return;
  }

  if (
    activeSocket &&
    activeSocketPairingId === pairingId &&
    (activeSocket.readyState === WebSocket.CONNECTING || activeSocket.readyState === WebSocket.OPEN)
  ) {
    return;
  }

  if (activeSocketPairingId && activeSocketPairingId !== pairingId) {
    teardownSocket();
  }

  openSocket(pairingId);
}

function subscribeSocket(pairingId: string, subscriber: SocketSubscriber) {
  socketSubscribers.add(subscriber);
  subscriber.onStatus(activeSocketStatus);
  ensureSocket(pairingId);

  return () => {
    socketSubscribers.delete(subscriber);

    if (socketSubscribers.size === 0) {
      teardownSocket();
    }
  };
}

export async function resolveCandidateChatThread(candidateId: string) {
  const response = await api.post("/chat/threads/resolve-candidate", {
    candidate_id: candidateId,
  });
  return extractThread(response.data);
}

export function useChatThreads() {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace =
    user?.role === UserRole.ETHIOPIAN_AGENT || user?.role === UserRole.FOREIGN_AGENT;

  return useQuery({
    queryKey: chatKeys.threads(activePairingId),
    queryFn: async () => {
      const response = await api.get("/chat/threads");
      return extractThreadList(response.data);
    },
    enabled: Boolean(user) && (!requiresWorkspace || (isPairingReady && Boolean(activePairingId))),
    staleTime: 20_000,
    refetchInterval: 30_000,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      return shouldRetryChatQuery(failureCount, error, 2);
    },
  });
}

export function useWorkspaceChatThread() {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);

  return useQuery({
    queryKey: chatKeys.workspaceThread(activePairingId),
    queryFn: async () => {
      const response = await api.post("/chat/threads/resolve-workspace", {});
      return extractThread(response.data);
    },
    enabled: Boolean(user) && isPairingReady && Boolean(activePairingId),
    staleTime: 20_000,
    refetchInterval: 30_000,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => shouldRetryChatQuery(failureCount, error, 1),
  });
}

export function useCandidateChatThread(candidateId?: string, options?: { enabled?: boolean }) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const enabled = options?.enabled ?? true;

  return useQuery({
    queryKey: chatKeys.candidateThread(activePairingId, candidateId),
    queryFn: async () => resolveCandidateChatThread(String(candidateId)),
    enabled: Boolean(candidateId) && enabled && Boolean(user) && isPairingReady && Boolean(activePairingId),
    staleTime: 20_000,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      return shouldRetryChatQuery(failureCount, error, 1);
    },
  });
}

export function useChatMessages(threadId?: string) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);

  const query = useInfiniteQuery({
    queryKey: chatKeys.messages(activePairingId, threadId),
    queryFn: async ({ pageParam }) => {
      const response = await api.get(`/chat/threads/${threadId}/messages`, {
        params: {
          limit: CHAT_PAGE_SIZE,
          ...(pageParam ? { cursor: pageParam } : {}),
        },
      });
      return extractMessagePage(response.data, user?.id);
    },
    enabled: Boolean(user) && Boolean(threadId) && isPairingReady && Boolean(activePairingId),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => (lastPage.hasMore ? lastPage.nextBefore || undefined : undefined),
    staleTime: 15_000,
    refetchInterval: 30_000,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => shouldRetryChatQuery(failureCount, error, 1),
  });

  const messages = React.useMemo(
    () => flattenMessagePages(query.data?.pages || []),
    [query.data?.pages]
  );

  return {
    ...query,
    messages,
    hasOlder: query.hasNextPage,
    loadOlder: query.fetchNextPage,
  };
}

export function useSendChatMessage(threadId?: string) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const queryClient = useQueryClient();

  return useMutation<ChatMessage, unknown, string, SendMessageContext>({
    mutationFn: async (body: string) => {
      if (!threadId) {
        throw new Error("Missing thread id");
      }

      const response = await api.post(`/chat/threads/${threadId}/messages`, {
        body,
      });

      return normalizeMessage((response.data as Record<string, unknown>).message || response.data, user?.id);
    },
    onMutate: async (rawBody) => {
      if (!threadId || !activePairingId) {
        return {
          optimisticMessageId: "",
        };
      }

      const body = rawBody.trim();
      const optimisticMessageId = `optimistic-${Date.now()}-${Math.random().toString(16).slice(2)}`;

      await queryClient.cancelQueries({ queryKey: chatKeys.messages(activePairingId, threadId) });

      const previousMessages = queryClient.getQueryData<InfiniteData<ChatMessagesPage>>(
        chatKeys.messages(activePairingId, threadId)
      );
      const previousThreads = queryClient.getQueryData<ChatThreadSummary[]>(
        chatKeys.threads(activePairingId)
      );

      const optimisticMessage: ChatMessage = {
        id: optimisticMessageId,
        thread_id: threadId,
        sender_user_id: String(user?.id || ""),
        sender_name: String(user?.full_name || "You"),
        sender_role: user?.role,
        body,
        created_at: new Date().toISOString(),
        is_mine: true,
      };

      queryClient.setQueryData<InfiniteData<ChatMessagesPage>>(
        chatKeys.messages(activePairingId, threadId),
        (current) => appendMessage(current, optimisticMessage)
      );

      queryClient.setQueryData<ChatThreadSummary[]>(chatKeys.threads(activePairingId), (threads) =>
        updateThreadLastMessage(threads, threadId, body)
      );

      return {
        previousMessages,
        optimisticMessageId,
        previousThreads,
      };
    },
    onError: (_error, _variables, context) => {
      if (!activePairingId || !threadId) {
        return;
      }

      if (context?.previousMessages) {
        queryClient.setQueryData(chatKeys.messages(activePairingId, threadId), context.previousMessages);
      }

      if (context?.previousThreads) {
        queryClient.setQueryData(chatKeys.threads(activePairingId), context.previousThreads);
      }

      if (isChatStorageUnavailableError(_error)) {
        toast.error("Chat is unavailable in this environment right now.");
        return;
      }

      toast.error("Failed to send message");
    },
    onSuccess: (message, _body, context) => {
      if (!activePairingId || !threadId) {
        return;
      }

      if (context?.optimisticMessageId) {
        queryClient.setQueryData<InfiniteData<ChatMessagesPage>>(
          chatKeys.messages(activePairingId, threadId),
          (current) => replaceOptimisticMessage(current, context.optimisticMessageId, message)
        );
      }
    },
    onSettled: async () => {
      if (!activePairingId || !threadId) {
        return;
      }

      await Promise.all([
        queryClient.invalidateQueries({ queryKey: chatKeys.messages(activePairingId, threadId) }),
        queryClient.invalidateQueries({ queryKey: chatKeys.threads(activePairingId) }),
        queryClient.invalidateQueries({ queryKey: chatKeys.summary(activePairingId) }),
      ]);
    },
  });
}

export function useMarkChatThreadRead(threadId?: string) {
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const queryClient = useQueryClient();

  return useMutation<unknown, unknown, void, MarkReadContext>({
    mutationFn: async () => {
      if (!threadId) {
        throw new Error("Missing thread id");
      }

      const response = await api.post(`/chat/threads/${threadId}/read`);
      return response.data;
    },
    onMutate: async () => {
      if (!activePairingId || !threadId) {
        return {};
      }

      await queryClient.cancelQueries({ queryKey: chatKeys.threads(activePairingId) });
      await queryClient.cancelQueries({ queryKey: chatKeys.summary(activePairingId) });

      const previousThreads = queryClient.getQueryData<ChatThreadSummary[]>(
        chatKeys.threads(activePairingId)
      );
      const previousSummary = queryClient.getQueryData<ChatSummary>(chatKeys.summary(activePairingId));
      const previousUnread = previousThreads?.find((thread) => thread.id === threadId)?.unread_count || 0;

      queryClient.setQueryData<ChatThreadSummary[]>(chatKeys.threads(activePairingId), (threads) =>
        applyReadState(threads, threadId, 0)
      );

      if (previousSummary) {
        queryClient.setQueryData<ChatSummary>(chatKeys.summary(activePairingId), {
          ...previousSummary,
          total_unread: Math.max(0, previousSummary.total_unread - previousUnread),
        });
      }

      return {
        previousThreads,
        previousSummary,
      };
    },
    onError: (_error, _variables, context) => {
      if (!activePairingId) {
        return;
      }

      if (context?.previousThreads) {
        queryClient.setQueryData(chatKeys.threads(activePairingId), context.previousThreads);
      }
      if (context?.previousSummary) {
        queryClient.setQueryData(chatKeys.summary(activePairingId), context.previousSummary);
      }

      if (isChatStorageUnavailableError(_error)) {
        toast.error("Chat is unavailable in this environment right now.");
        return;
      }

      toast.error("Failed to mark chat thread as read");
    },
    onSettled: async () => {
      if (!activePairingId) {
        return;
      }

      await Promise.all([
        queryClient.invalidateQueries({ queryKey: chatKeys.threads(activePairingId) }),
        queryClient.invalidateQueries({ queryKey: chatKeys.summary(activePairingId) }),
      ]);
    },
  });
}

export function useChatSummary() {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace =
    user?.role === UserRole.ETHIOPIAN_AGENT || user?.role === UserRole.FOREIGN_AGENT;

  const query = useQuery({
    queryKey: chatKeys.summary(activePairingId),
    queryFn: async () => {
      const response = await api.get<ChatSummaryResponse>("/chat/summary");
      return normalizeSummary(response.data.summary || response.data);
    },
    enabled: Boolean(user) && (!requiresWorkspace || (isPairingReady && Boolean(activePairingId))),
    staleTime: 20_000,
    refetchInterval: 30_000,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      return shouldRetryChatQuery(failureCount, error, 1);
    },
  });

  return {
    ...query,
    summary: query.data || {
      total_unread: 0,
      workspace_unread: 0,
      candidate_unread: 0,
      threads_with_unread: 0,
    },
    count: query.data?.total_unread || 0,
  };
}

export function useChatRealtime() {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const queryClient = useQueryClient();
  const [status, setStatus] = React.useState<ChatRealtimeStatus>("idle");

  React.useEffect(() => {
    if (!user || !activePairingId || !isPairingReady) {
      setStatus("idle");
      return;
    }

    const unsubscribe = subscribeSocket(activePairingId, {
      onStatus: (nextStatus) => {
        setStatus(nextStatus);
      },
      onEvent: (event) => {
        applySocketEventToCache(queryClient, activePairingId, event, user.id);
      },
    });

    return unsubscribe;
  }, [activePairingId, isPairingReady, queryClient, user]);

  return {
    status,
    isConnected: status === "connected",
    isReconnecting: status === "connecting" || status === "reconnecting",
    isDisconnected: status === "disconnected" || status === "idle",
  };
}
