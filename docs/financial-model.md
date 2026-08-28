# مدل مالی Hesab-keepnet

این سند مرجع رسمی قوانین مالی سیستم است. هر تغییری در منطق مالی باید ابتدا در
این سند بازتاب داده شود.

---

## ۱. اصل بنیادین: Ledger به عنوان Source of Truth

```text
اسناد کسب‌وکاری (Sale / SalePayment / Expense / Transfer)
        ↓  تولید اتمیک
دفتر تراکنش‌ها (transactions) ← تنها منبع حرکت پول بین حساب‌ها
        ↓  مشتق‌سازی
مانده‌ها و گزارش‌ها
```

* جدول `transactions` تنها منبع حقیقتِ حرکت پول است.
* هیچ فیلد Balance ذخیره‌شده‌ای وجود ندارد (تست `TestNoStoredBalanceColumn` این را تضمین می‌کند).
* اسناد کسب‌وکاری «معنا» را نگه می‌دارند؛ دفتر «اثر پولی» را.

## ۲. نمایش پول

| موضوع | قانون |
|---|---|
| نوع داده | `int64` — بدون هیچ استثنایی |
| TOMAN | تومان صحیح |
| USD | **سنت** (برای جلوگیری از اعشار) |
| Float | مطلقاً ممنوع؛ تست `TestNoFloatingPointInMoneyPackages` مخالفت float32/float64 در پکیج‌های مالی را می‌شکند |
| Magic number | مبالغ فقط از ورودی کاربر/تست می‌آیند |

## ۳. Currency

* دو ارز مجاز: `TOMAN`، `USD` (CHECK در سطح DB).
* ارز هر ردیف دفتر باید با ارز حساب بانکی برابر باشد → **Trigger** `trg_transactions_currency_match` در DB + validation در Service.
* انتقال بین دو حساب غیرهم‌ارز ممنوع.
* هیچ گزارشی دو ارز را جمع نمی‌کند؛ خروجی‌ها همیشه per-currency هستند.

## ۴. مانده حساب بانکی (Derived)

```text
Balance(account) = InitialBalance
                 + Σ transactions(INCOME, TRANSFER_IN)
                 − Σ transactions(EXPENSE, TRANSFER_OUT)
                 [deleted_at IS NULL]
```

پیاده‌سازی: `BankAccountService.Balance`. هیچ مسیر CRUD برای نوشتن مانده وجود ندارد.

## ۵. فروش (Sale)

```text
Sale (total_amount, currency, sold_at, customer_name?)
 └── SalePayment (gateway: ZARINPAL|CARD_TO_CARD|SUPPORT, amount, bank_account_id, gateway_ref?)
      └── Ledger INCOME row (اتمیک)
```

قوانین:

* `amount > 0` برای کل فروش و هر پرداخت (CHECK در DB).
* ارز پرداخت = ارز حساب = ارز فروش (Service + Trigger).
* `Σ Payments ≤ TotalAmount` — فروش بیش از مبلغ کل ممنوع (پرداخت جزئی مجاز).
* **TotalAmount ذخیره می‌شود** چون «قیمت توافق‌شده» یک واقعیت کسب‌وکاری مستقل از پرداخت‌هاست؛ Drift آن ناممکن است چون مقایسه با پرداخت‌ها فقط یک‌طرفه است (`paid ≤ total`) و PaidAmount همیشه `SUM` زنده است.
* Status ذخیره نمی‌شود: `UNPAID / PARTIAL / PAID` از نسبت `PaidAmount ÷ TotalAmount` مشتق می‌شود (`services.StatusOf`).

## ۶. هزینه (Expense)

```text
Expense (category_id, bank_account_id?, amount, currency, occurred_at)
 └── Ledger EXPENSE row — فقط وقتی bank_account_id تعیین شده باشد
```

* ماهیت Business/Personal **تنها از `Category.Type`** خوانده می‌شود؛ فیلد موازی وجود ندارد.
* هزینه BUSINESS باید حساب بانکی داشته باشد (Service rule).
* هزینه PERSONAL نقدی (بدون حساب) مجاز است → ردیابی آماری بدون اثر بانکی.
* هزینه PERSONALِ متصل به حساب، موجودی حساب را کم می‌کند ولی هرگز وارد سود نمی‌شود.

## ۷. انتقال (Transfer)

* یک سند → دقیقاً **دو** ردیف دفتر: `TRANSFER_OUT` (مبدأ) و `TRANSFER_IN` (مقصد).
* `from ≠ to` (CHECK در DB)، `currency(from) == currency(to) == currency(transfer)`، `amount > 0`.
* Transfer در Revenue/Expense/Profit ظاهر نمی‌شود (تست‌شده).
* ثبت/حذف کاملاً atomic است (تست تزریق خطا بین دو ردیف).

## ۸. نماینده‌ها (Representative Ledger)

```text
RepresentativeTransaction(direction: DEBIT | CREDIT, amount > 0, currency = rep.currency)
```

چرا Direction به جای Amount علامت‌دار؟

1. خوانایی گزارش صورت‌حساب (هر ردیف معنای خودش را دارد).
2. CHECK constraint روی enum ممکن است؛ برای عدد علامت‌دار باید `CHECK(amount<>0)` و تفکیک مثبت/منفی در کوئری‌ها انجام شود که خطاخیزتر است.
3. جمع‌ها با `GROUP BY direction` deterministic و index-friendly هستند.

مانده (per currency):

```text
Balance = TotalDebit − TotalCredit
Balance > 0  ⟹  نماینده به کسب‌وکار بدهکار است
Balance < 0  ⟹  کسب‌وکار به نماینده بدهکار است
```

* ارز هر نماینده واحد است (فیلد `currency`)؛ تراکنش با ارز دیگر رد می‌شود.
* کمیسیون در Basis Point ذخیره می‌شود: `commission_percent_bps ∈ [0, 10000]` (CHECK)؛ `1250 = 12.5%`.
* دفتر نماینده مستقل از دفتر بانکی است و در Net Profit ورود ندارد.

## ۹. سود خالص (Cash Basis)

```text
Net Profit(period, per currency) =
    Σ transactions(type = INCOME, occurred_at ∈ period)     ← پول واقعاً دریافت‌شده
  − Σ Expenses(category.type = BUSINESS, occurred_at ∈ period)
```

* Personal Expense در هیچ شکل از فرمول حذف شده است.
* Transfer ذاتاً خارج است (typeهای TRANSFER_* محاسبه نمی‌شوند).
* پیاده‌سازی: `ReportingService.NetProfit`.

## ۱۰. حذف اسناد — Soft Delete و تاریخچه

سیستم از **Soft Delete یکپارچه** استفاده می‌کند:

* حذف سند ⟹ soft-delete سند + همه ردیف‌های دفتر مرتبطش، در همان تراکنش DB.
* ردیف‌ها از DB حذف **نمی‌شوند** (`deleted_at` ست می‌شود) → حفظ ردپای audit.
* تمام محاسبات (مانده، سود، گزارش‌ها) به‌صورت یکدست ردیف‌های deleted را کنار می‌گذارند؛ پس هیچ محاسبه‌ای نیمه‌کاره تغییر نمی‌کند.
* Hard Delete در سطح Service وجود ندارد؛ FKها RESTRICT هستند تا حتی خطای برنامه‌نویس هم نتواند تاریخچه را پاک کند.

**گزینه بررسی‌شده اما به فاز آینده موکول‌شده:** الگوی VOID/REVERSAL (ثبت ردیف معکوس به‌جای حذف). مزیت آن حفظ جریان کامل در خروجی‌های دفتری است، اما پیچیدگی همه گزارش‌ها را بالا می‌برد؛ برای v1 تک‌کاربره، Soft Delete یکپارچه + AuditLog کافی و صحیح است. مهاجرت آینده به REVERSAL ساختار فعلی را نمی‌شکند (ردیف‌ها immutable هستند).

## ۱۱. Audit Trail

* هر عملیات ایجاد/حذف مالی یک ردیف `audit_logs` **داخل همان تراکنش DB** می‌نویسد → اگر عملیات rollback شود، لاگش هم هست یا هیچ‌کدام نیستند.
* `AuditLog` append-only است (بدون update/delete در API).

## ۱۲. Invariants و محل enforce شدن

| # | Invariant | لایه |
|---|---|---|
| 1 | `amount > 0` | CHECK (DB) + Service |
| 2 | ارز معتبر TOMAN/USD | CHECK (DB) + Service |
| 3 | ارز ردیف دفتر = ارز حساب | **Trigger (DB)** + Service |
| 4 | هر ردیف دفتر دقیقاً یک سند منبع مطابق type دارد | CHECK ترکیبی (DB) |
| 5 | `transfer.from ≠ transfer.to` | CHECK (DB) + Service |
| 6 | ارز یکسان برای مبدأ/مقصد/انتقال | Service (+ تست) |
| 7 | Business Expense ⟹ حساب بانکی الزامی | Service + تست |
| 8 | Category.Type منبع یگانه ماهیت هزینه | طراحی schema (بدون فیلد موازی) |
| 9 | Type فرزند = Type والد در دسته‌ها | Service + تست |
| 10 | `commission_percent_bps ∈ [0,10000]` | CHECK (DB) + Service |
| 11 | `Σ payments ≤ sale.total` | Service + تست |
| 12 | Username unique (بین فعال‌ها) | Unique partial index |
| 13 | NationalCode unique (بین فعال‌ها) | Unique partial index |
| 14 | عدم وجود ستون Balance قابل ویرایش | طراحی schema + تست |

## ۱۳. Seed

* ۱۲ دسته پیش‌فرض (۵ Business + ۷ Personal) — idempotent با شرط NOT EXISTS منطقی.
* Admin فقط در development و فقط با `ADMIN_USERNAME` + `ADMIN_PASSWORD` ساخته می‌شود؛ password با Argon2id هش می‌شود و re-seed پسورد موجود را بازنویسی نمی‌کند.
* کل seed داخل یک `BEGIN IMMEDIATE` transaction است.
