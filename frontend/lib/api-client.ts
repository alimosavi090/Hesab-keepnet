import type { ApiEnvelope } from "@/types/api";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(code: string, message: string, status: number, cause?: unknown) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    if (cause !== undefined) this.cause = cause;
  }
}

export interface RequestOptionsParams {
  [key: string]: unknown;
}

type QueryParams = Record<string, string | number | boolean | undefined>;

type RequestOptions = {
  query?: QueryParams;
  signal?: AbortSignal;
};

function csrfToken(): string | undefined {
  if (typeof document === "undefined") return undefined;
  const match = document.cookie.match(/(?:^|;\s*)hesab_csrf=([^;]+)/);
  return match?.[1];
}

function buildUrl(path: string, options?: RequestOptions): string {
  const url = new URL(`${BASE_URL}${path}`);
  if (options?.query) {
    for (const [key, value] of Object.entries(options.query)) {
      if (value !== undefined && value !== "") url.searchParams.set(key, String(value));
    }
  }
  return url.toString();
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  options?: RequestOptions
): Promise<T> {
  let response: Response;

  try {
    response = await fetch(buildUrl(path, options), {
      method,
      credentials: "include",
      headers: {
        ...(body !== undefined ? { "Content-Type": "application/json" } : {}),
        ...(["POST", "PUT", "PATCH", "DELETE"].includes(method)
          ? { "X-CSRF-Token": csrfToken() ?? "" }
          : {}),
      },
      body: body !== undefined ? JSON.stringify(body) : undefined,
      signal: options?.signal,
      cache: "no-store",
    });
  } catch (cause) {
    throw new ApiError("NETWORK_ERROR", "ارتباط با سرور برقرار نشد.", 0, cause);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  let envelope: ApiEnvelope<T>;
  try {
    envelope = (await response.json()) as ApiEnvelope<T>;
  } catch (cause) {
    throw new ApiError(
      "INVALID_RESPONSE",
      "پاسخ نامعتبری از سرور دریافت شد.",
      response.status,
      cause
    );
  }

  if (!response.ok || !envelope.success || envelope.error !== null) {
    throw new ApiError(
      envelope.error?.code ?? "UNKNOWN_ERROR",
      envelope.error?.message ?? "خطای ناشناخته‌ای رخ داد.",
      response.status
    );
  }

  return envelope.data as T;
}

export const apiClient = {
  get: <T>(path: string, params?: QueryParams, signal?: AbortSignal) =>
    request<T>("GET", path, undefined, { query: params, signal }),
  post: <T>(path: string, body?: unknown) => request<T>("POST", path, body),
  put: <T>(path: string, body?: unknown) => request<T>("PUT", path, body),
  patch: <T>(path: string, body?: unknown) => request<T>("PATCH", path, body),
  delete: <T>(path: string) => request<T>("DELETE", path),
};

export async function downloadFile(path: string, fallbackName: string): Promise<void> {
  const response = await fetch(`${BASE_URL}${path}`, {
    credentials: "include",
    cache: "no-store",
  });
  if (!response.ok) {
    throw new ApiError("DOWNLOAD_FAILED", "دانلود فایل ناموفق بود.", response.status);
  }
  const blob = await response.blob();
  const disposition = response.headers.get("Content-Disposition") ?? "";
  const match = disposition.match(/filename=([\w.-]+)/);
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = match?.[1] ?? fallbackName;
  link.click();
  URL.revokeObjectURL(link.href);
}
