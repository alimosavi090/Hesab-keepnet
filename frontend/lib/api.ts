import { downloadFile } from "@/lib/api-client";
import type {
  AccountBalance,
  BackupFile,
  BackupListResponse,
  Note,
  NoteEntityType,
  PublicUser,
  Transfer,
  BankAccount,
  Category,
  CategoryType,
  Currency,
  DailyPoint,
  DashboardSummary,
  Expense,
  Gateway,
  LedgerItem,
  LedgerType,
  Paged,
  ReportOverview,
  Reminder,
  RepDirection,
  Representative,
  RepresentativeBalance,
  RepresentativeTransaction,
  Sale,
  SaleListItem,
} from "@/types/api";

export interface PagedQuery {
  page?: number;
  page_size?: number;
}

export interface DateRangeQuery {
  from?: string;
  to?: string;
}

import { apiClient } from "@/lib/api-client";

type QueryParams = Record<string, string | number | boolean | undefined>;

async function apiGet<T>(path: string, params?: object): Promise<T> {
  return apiClient.get<T>(path, params as QueryParams | undefined);
}
async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  return apiClient.post<T>(path, body);
}
async function apiPatch<T>(path: string, body?: unknown): Promise<T> {
  return apiClient.patch<T>(path, body);
}
async function apiDelete(path: string): Promise<void> {
  await apiClient.delete(path);
}

async function apiPaged<T>(path: string, params: object): Promise<Paged<T>> {
  return apiClient.get<Paged<T>>(path, params as QueryParams);
}

export const authApi = {
  login: (username: string, password: string) =>
    apiPost<PublicUser>("/api/v1/auth/login", { username, password }),
  logout: () => apiPost("/api/v1/auth/logout"),
  me: () => apiGet<PublicUser>("/api/v1/auth/me"),
};

export const reportsApi = {
  summary: (range: DateRangeQuery) => apiGet<ReportOverview>(`/api/v1/reports/summary`, range),
  download: (dataset: "expenses" | "sales" | "transactions", range: DateRangeQuery) => {
    const search = new URLSearchParams();
    search.set("dataset", dataset);
    for (const [key, value] of Object.entries(range ?? {})) {
      if (value) search.set(key, String(value));
    }
    return downloadFile(`/api/v1/reports/export.csv?${search.toString()}`, `${dataset}.csv`);
  },
};

export const bankAccountsApi = {
  list: (includeInactive = false) =>
    apiGet<BankAccount[]>(`/api/v1/bank-accounts`, { include_inactive: includeInactive ? "true" : undefined }),
  balance: (id: number) => apiGet<AccountBalance>(`/api/v1/bank-accounts/${id}/balance`),
  create: (input: { name: string; bank_name: string; card_number?: string; currency: Currency; initial_balance: number; description?: string }) =>
    apiPost<BankAccount>(`/api/v1/bank-accounts`, input),
  setActive: (id: number, is_active: boolean) =>
    apiPatch<unknown>(`/api/v1/bank-accounts/${id}`, { is_active }),
  update: (id: number, input: { name?: string; bank_name?: string; card_number?: string; description?: string; initial_balance?: number }) =>
    apiPatch<BankAccount>(`/api/v1/bank-accounts/${id}/edit`, input),
};

export const categoriesApi = {
  list: (type?: CategoryType) => apiGet<Category[]>(`/api/v1/categories`, { type }),
  create: (input: { name: string; type: CategoryType; parent_id?: number }) =>
    apiPost<Category>(`/api/v1/categories`, input),
  deactivate: (id: number) => apiDelete(`/api/v1/categories/${id}`),
};

export const expensesApi = {
  list: (params: PagedQuery & DateRangeQuery & { currency?: Currency; category_id?: number; bank_account_id?: number; type?: CategoryType }) =>
    apiPaged<Expense>(`/api/v1/expenses`, params),
  create: (input: { category_id: number; bank_account_id?: number; amount: number; currency: Currency; occurred_at: string; description?: string }) =>
    apiPost<Expense>(`/api/v1/expenses`, input),
  update: (id: number, input: { category_id?: number; amount?: number; occurred_at?: string; description?: string }) =>
    apiPatch<Expense>(`/api/v1/expenses/${id}`, input),
  remove: (id: number) => apiDelete(`/api/v1/expenses/${id}`),
};

export const salesApi = {
  list: (params: PagedQuery & DateRangeQuery & { currency?: Currency; gateway?: Gateway }) =>
    apiPaged<SaleListItem>(`/api/v1/sales`, params),
  get: (id: number) => apiGet<{ sale: Sale; paid_amount: number; status: string }>(`/api/v1/sales/${id}`),
  create: (input: {
    total_amount: number;
    currency: Currency;
    sold_at: string;
    customer_name?: string;
    payments: Array<{ bank_account_id: number; gateway: Gateway; amount: number; paid_at: string; gateway_ref?: string }>;
  }) => apiPost<Sale>(`/api/v1/sales`, input),
  update: (id: number, input: { total_amount?: number; customer_name?: string; sold_at?: string }) =>
    apiPatch<Sale>(`/api/v1/sales/${id}`, input),
  remove: (id: number) => apiDelete(`/api/v1/sales/${id}`),
};

export const transfersApi = {
  list: (params: PagedQuery & DateRangeQuery) => apiPaged<Transfer>(`/api/v1/transfers`, params),
  create: (input: { from_account_id: number; to_account_id: number; amount: number; currency: Currency; transferred_at: string; description?: string }) =>
    apiPost<Transfer>(`/api/v1/transfers`, input),
  remove: (id: number) => apiDelete(`/api/v1/transfers/${id}`),
};

export const representativesApi = {
  list: (includeInactive = false) =>
    apiGet<Representative[]>(`/api/v1/representatives`, { include_inactive: includeInactive ? "true" : undefined }),
  create: (input: { full_name: string; phone: string; national_code?: string; currency: Currency; start_date: string }) =>
    apiPost<Representative>(`/api/v1/representatives`, input),
  balance: (id: number) => apiGet<RepresentativeBalance>(`/api/v1/representatives/${id}/balance`),
  transactions: (id: number, params: PagedQuery) => apiPaged<RepresentativeTransaction>(`/api/v1/representatives/${id}/transactions`, params),
  recordTransaction: (
    id: number,
    input: { direction: RepDirection; amount: number; occurred_at: string; bank_account_id?: number; description?: string },
  ) =>
    apiPost<RepresentativeTransaction>(`/api/v1/representatives/${id}/transactions`, input),
  removeTransaction: (txnId: number) => apiDelete(`/api/v1/rep-transactions/${txnId}`),
};

export const remindersApi = {  list: () => apiGet<Reminder[]>(`/api/v1/reminders`),
  upcoming: (days = 7) => apiGet<Reminder[]>(`/api/v1/reminders/upcoming`, { days }),
  create: (input: { title: string; due_date: string; description?: string; repeat_interval?: "NONE" | "MONTHLY" | "YEARLY" }) =>
    apiPost<Reminder>(`/api/v1/reminders`, input),
  setDone: (id: number, is_done: boolean) => apiPatch<unknown>(`/api/v1/reminders/${id}`, { is_done }),
  remove: (id: number) => apiDelete(`/api/v1/reminders/${id}`),
};

export const transactionsApi = {
  feed: (params: PagedQuery & DateRangeQuery & { bank_account_id?: number; type?: LedgerType; currency?: Currency }) =>
    apiPaged<LedgerItem>(`/api/v1/transactions`, params),
};

export const dashboardApi = {
  summary: (params?: DateRangeQuery) => apiGet<DashboardSummary>(`/api/v1/dashboard/summary`, params),
  chart: (days: number, currency: Currency) => apiGet<DailyPoint[]>(`/api/v1/dashboard/chart`, { days, currency }),
};

export const notesApi = {
  list: (params: PagedQuery & {
    entity_type?: NoteEntityType;
    entity_id?: number;
    pinned?: boolean;
    tag?: string;
    q?: string;
  }) =>
    apiPaged<Note>(`/api/v1/notes`, params),
  create: (input: {
    entity_type: NoteEntityType;
    entity_id?: number;
    body: string;
    tags?: string[];
    pinned?: boolean;
  }) => apiPost<Note>(`/api/v1/notes`, input),
  update: (id: number, input: { body?: string; tags?: string[]; pinned?: boolean }) =>
    apiPatch<Note>(`/api/v1/notes/${id}`, input),
  remove: (id: number) => apiDelete(`/api/v1/notes/${id}`),
};

export const backupsApi = {
  list: () => apiGet<BackupListResponse>(`/api/v1/backups`),
  create: () => apiPost<BackupFile>(`/api/v1/backups`),
  download: (name: string) =>
    downloadFile(`/api/v1/backups/${encodeURIComponent(name)}`, name),
  remove: (name: string) => apiDelete(`/api/v1/backups/${encodeURIComponent(name)}`),
};
