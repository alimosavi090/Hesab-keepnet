export type Currency = "RIAL" | "USD";

export interface ApiErrorInfo {
  code: string;
  message: string;
}

export interface ApiEnvelope<T> {
  success: boolean;
  data: T | null;
  error: ApiErrorInfo | null;
}

export interface HealthInfo {
  status: "ok" | "degraded";
  database: "up" | "down";
  environment: string;
  version: string;
}
export type CategoryType = "BUSINESS" | "PERSONAL";
export type Gateway = "ZARINPAL" | "CARD_TO_CARD" | "SUPPORT";
export type LedgerType = "INCOME" | "EXPENSE" | "TRANSFER_IN" | "TRANSFER_OUT";
export type RepDirection = "DEBIT" | "CREDIT";
export type RepeatInterval = "NONE" | "MONTHLY" | "YEARLY";
export type SaleStatus = "UNPAID" | "PARTIAL" | "PAID";

export interface BankAccount {
  id: number;
  name: string;
  bank_name: string;
  card_number: string | null;
  currency: Currency;
  initial_balance: number;
  description: string | null;
  is_active: boolean;
}

export interface AccountBalance {
  account_id: number;
  currency: Currency;
  initial_balance: number;
  incoming: number;
  outgoing: number;
  balance: number;
}

export interface Category {
  id: number;
  name: string;
  type: CategoryType;
  parent_id: number | null;
  is_active: boolean;
}

export interface SalePayment {
  id?: number;
  bank_account_id: number;
  gateway: Gateway;
  amount: number;
  paid_at: string;
  gateway_ref?: string | null;
  description?: string | null;
}

export interface Sale {
  id: number;
  total_amount: number;
  currency: Currency;
  sold_at: string;
  customer_name: string | null;
  description: string | null;
  payments?: SalePayment[];
}

export interface SaleListItem extends Sale {
  paid_amount: number;
  status: SaleStatus;
}

export interface Expense {
  id: number;
  category_id: number;
  bank_account_id: number | null;
  amount: number;
  currency: Currency;
  occurred_at: string;
  description: string | null;
  category?: Category;
}

export interface Transfer {
  id: number;
  from_account_id: number;
  to_account_id: number;
  amount: number;
  currency: Currency;
  transferred_at: string;
  description: string | null;
  from_account?: Pick<BankAccount, "id" | "name">;
  to_account?: Pick<BankAccount, "id" | "name">;
}

export interface Representative {
  id: number;
  full_name: string;
  phone: string;
  email: string | null;
  national_code: string | null;
  currency: Currency;
  start_date: string;
  is_active: boolean;
}

export interface RepresentativeTransaction {
  id: number;
  representative_id: number;
  direction: RepDirection;
  amount: number;
  currency: Currency;
  occurred_at: string;
  description: string | null;
  bank_account_id: number | null;
  ledger_transaction_id: number | null;
  bank_account?: Pick<BankAccount, "id" | "name" | "bank_name"> | null;
}

export interface RepresentativeDebt {
  representative_id: number;
  full_name: string;
  currency: Currency;
  debt: number;
}

export interface RepresentativeBalance {
  currency: Currency;
  total_debit: number;
  total_credit: number;
  balance: number;
}

export interface Reminder {
  id: number;
  title: string;
  description: string | null;
  due_date: string;
  repeat_interval: RepeatInterval;
  is_done: boolean;
  completed_at: string | null;
}

export interface LedgerItem {
  id: number;
  bank_account_id: number;
  account_name: string;
  type: LedgerType;
  amount: number;
  currency: Currency;
  occurred_at: string;
  description: string | null;
  source_type: string;
  source_id: number | null;
}

export interface ProfitRow {
  currency: Currency;
  sales: number;
  business_expense: number;
  net_profit: number;
}

export interface GatewayTotal {
  gateway: Gateway;
  currency: Currency;
  total: number;
}

export interface TotalRow {
  currency: Currency;
  total: number;
}

export interface ExpenseSplit {
  business: TotalRow[];
  personal: TotalRow[];
}

export interface DashboardSummary {
  profit: ProfitRow[];
  sales_by_gateway: GatewayTotal[];
  expenses: ExpenseSplit;
  banks: (AccountBalance & { name?: string })[];
  recent: LedgerItem[];
  reminders: Reminder[];
  rep_debts?: RepresentativeDebt[];
}

export interface CategoryExpenseRow {
  category_name: string;
  category_type: CategoryType;
  currency: Currency;
  total: number;
}

export interface RepSettlementRow {
  currency: Currency;
  total: number;
  count: number;
}

export interface ReportOverview {
  profit: ProfitRow[];
  gateways: GatewayTotal[];
  expenses_by_category: CategoryExpenseRow[];
  rep_settlements: RepSettlementRow[];
  rep_debts?: RepresentativeDebt[];
  expense_split: ExpenseSplit;
}

export type NoteEntityType = "REPRESENTATIVE" | "SALE" | "BANK_ACCOUNT" | "JOURNAL";

export interface Note {
  id: number;
  entity_type: NoteEntityType;
  entity_id: number | null;
  body: string;
  tags: string;
  pinned: boolean;
  created_at: string;
  updated_at: string;
}

export interface BackupFile {
  name: string;
  size_bytes: number;
  created_at: string;
  is_auto: boolean;
}

export interface BackupListResponse {
  items: BackupFile[];
  last_auto_at: string | null;
  interval_hours: number;
}

export interface DailyPoint {
  date: string;
  sales: number;
  business_expense: number;
}

export interface PublicUser {
  id: number;
  username: string;
  display_name: string;
  role: string;
}

export interface Paged<T> {
  items: T[];
  meta: { total: number; page: number; page_size: number };
}
