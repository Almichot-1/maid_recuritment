export type CandidateDraftDocumentType = "passport" | "photo" | "video";

export type CandidateDraftFormValues = {
  full_name?: string;
  nationality?: string;
  date_of_birth?: string;
  age?: number | string;
  place_of_birth?: string;
  religion?: string;
  marital_status?: string;
  children_count?: number | string;
  education_level?: string;
  experience_years?: number | string;
  skills?: string[];
  languages?: Array<{ language: string; proficiency: string }>;
};

export type CandidateDraftSnapshot = {
  formValues: Partial<CandidateDraftFormValues>;
  documents: Partial<
    Record<
      CandidateDraftDocumentType,
      {
        name: string;
        size: number;
        type: string;
      }
    >
  >;
  updatedAt: number;
};

const SNAPSHOT_KEY = "candidate-new-draft:v1";
const DB_NAME = "maid-recruitment-drafts";
const STORE_NAME = "candidate-form-files";

function canUseBrowserStorage() {
  return typeof window !== "undefined";
}

function readSnapshot(): CandidateDraftSnapshot | null {
  if (!canUseBrowserStorage()) {
    return null;
  }

  const raw = window.sessionStorage.getItem(SNAPSHOT_KEY);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as CandidateDraftSnapshot;
    return {
      formValues: parsed.formValues || {},
      documents: parsed.documents || {},
      updatedAt: parsed.updatedAt || Date.now(),
    };
  } catch {
    return null;
  }
}

function writeSnapshot(snapshot: CandidateDraftSnapshot) {
  if (!canUseBrowserStorage()) {
    return;
  }

  window.sessionStorage.setItem(SNAPSHOT_KEY, JSON.stringify(snapshot));
}

function buildSnapshot(
  partial: Partial<CandidateDraftSnapshot>,
): CandidateDraftSnapshot {
  const current = readSnapshot();

  return {
    formValues: partial.formValues ?? current?.formValues ?? {},
    documents: partial.documents ?? current?.documents ?? {},
    updatedAt: Date.now(),
  };
}

function getDraftFileKey(documentType: CandidateDraftDocumentType) {
  return `${SNAPSHOT_KEY}:${documentType}`;
}

function openDraftDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    if (!canUseBrowserStorage() || !("indexedDB" in window)) {
      reject(new Error("indexedDB unavailable"));
      return;
    }

    const request = window.indexedDB.open(DB_NAME, 1);

    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME);
      }
    };

    request.onerror = () => {
      reject(request.error || new Error("failed to open draft database"));
    };

    request.onsuccess = () => {
      resolve(request.result);
    };
  });
}

function runFileTransaction<T>(
  mode: IDBTransactionMode,
  callback: (store: IDBObjectStore, resolve: (value: T) => void) => void,
): Promise<T> {
  return new Promise(async (resolve, reject) => {
    try {
      const db = await openDraftDatabase();
      const transaction = db.transaction(STORE_NAME, mode);
      const store = transaction.objectStore(STORE_NAME);

      transaction.onabort = () => {
        reject(transaction.error || new Error("draft file transaction aborted"));
      };
      transaction.onerror = () => {
        reject(transaction.error || new Error("draft file transaction failed"));
      };
      transaction.oncomplete = () => {
        db.close();
      };

      callback(store, resolve);
    } catch (error) {
      reject(error);
    }
  });
}

export function readCandidateDraftSnapshot() {
  return readSnapshot();
}

export function saveCandidateDraftFormValues(
  formValues: Partial<CandidateDraftFormValues>,
) {
  writeSnapshot(
    buildSnapshot({
      formValues,
    }),
  );
}

export function saveCandidateDraftDocumentMeta(
  documentType: CandidateDraftDocumentType,
  file: File | null,
) {
  const current = buildSnapshot({});
  const nextDocuments = { ...current.documents };

  if (file) {
    nextDocuments[documentType] = {
      name: file.name,
      size: file.size,
      type: file.type,
    };
  } else {
    delete nextDocuments[documentType];
  }

  writeSnapshot(
    buildSnapshot({
      formValues: current.formValues,
      documents: nextDocuments,
    }),
  );
}

export async function saveCandidateDraftFile(
  documentType: CandidateDraftDocumentType,
  file: File,
) {
  await runFileTransaction<void>("readwrite", (store, resolve) => {
    store.put(file, getDraftFileKey(documentType));
    resolve();
  });
}

export async function loadCandidateDraftFile(
  documentType: CandidateDraftDocumentType,
) {
  return runFileTransaction<File | null>("readonly", (store, resolve) => {
    const request = store.get(getDraftFileKey(documentType));
    request.onsuccess = () => {
      resolve((request.result as File | undefined) ?? null);
    };
    request.onerror = () => {
      resolve(null);
    };
  });
}

export async function clearCandidateDraftFile(
  documentType: CandidateDraftDocumentType,
) {
  await runFileTransaction<void>("readwrite", (store, resolve) => {
    store.delete(getDraftFileKey(documentType));
    resolve();
  });
}

export async function clearCandidateDraft() {
  if (canUseBrowserStorage()) {
    window.sessionStorage.removeItem(SNAPSHOT_KEY);
  }

  await Promise.all([
    clearCandidateDraftFile("passport"),
    clearCandidateDraftFile("photo"),
    clearCandidateDraftFile("video"),
  ]).catch(() => {
    // Ignore cleanup issues for temporary drafts.
  });
}
