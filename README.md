# Hesab-keepnet — سیستم حسابداری فروش VPN

نرم‌افزار حسابداری اختصاصی برای کسب‌وکار فروش VPN، شامل مدیریت فروش، هزینه‌ها
(کسب‌وکار/شخصی)، حساب‌های بانکی چندارزی، نماینده‌ها و گزارش‌های مالی.

## معماری کلی

| بخش | تکنولوژی |
|---|---|
| Backend | Go · Gin · GORM · SQLite (WAL) · REST |
| Frontend | Next.js (App Router) · TypeScript Strict · Tailwind CSS v4 · shadcn/ui · TanStack Query · Framer Motion · Vazirmatn |

ساختار مونورپو:

```text
backend/     سرویس REST روی پورت پیش‌فرض 8080
frontend/    اپلیکیشن Next.js روی پورت پیش‌فرض 3000
```

## پیش‌نیازها

- Go 1.25+
- Node.js 20+ و npm

## راه‌اندازی Backend

```bash
cd backend
cp .env.example .env          # مقدار SESSION_SECRET را حتماً تغییر دهید
go mod tidy
go run ./cmd/server           # اجرای توسعه
```

سرور روی `http://localhost:8080` بالا می‌آید.

### متغیرهای محیطی Backend (`.env`)

| متغیر | پیش‌فرض | توضیح |
|---|---|---|
| `APP_ENV` | `development` | `development` / `test` / `production` |
| `APP_PORT` | `8080` | پورت HTTP |
| `DATABASE_PATH` | `./data/accounting.db` | مسیر فایل SQLite (پوشه خودکار ساخته می‌شود) |
| `CORS_ORIGIN` | `http://localhost:3000` | لیست Originهای مجاز، جداشده با کاما |
| `SESSION_SECRET` | — (توسعه: تصادفی موقت) | حداقل ۱۶ کاراکتر؛ در production اجباری است |
| `COOKIE_SECURE` | `false` (production: `true`) | کوکی Secure برای احراز هویت |
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | خالی (فقط development) | ساخت خودکار کاربر admin در اولین Seed |

در حالت development اگر `SESSION_SECRET` تنظیم نشده باشد، یک secret موقت تولید
می‌شود و در لاگ هشدار داده می‌شود.

## راه‌اندازی Frontend

```bash
cd frontend
npm install
cp .env.example .env.local    # NEXT_PUBLIC_API_URL را در صورت نیاز تنظیم کنید
npm run dev                   # اجرای توسعه
```

اپلیکیشن روی `http://localhost:3000` در دسترس است.

### متغیرهای محیطی Frontend (`.env.local`)

| متغیر | پیش‌فرض | توضیح |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | آدرس پایه API بک‌اند |

## دستورات مفید

```bash
# Backend
go test ./...                 # تست‌ها
go build ./...                # کامپایل
go run ./cmd/server           # اجرا

# Frontend
npm run lint                  # ESLint
npm run build                 # بیلد پروداکشن
npm run start                 # اجرای بیلد پروداکشن
```

## Health Check

```bash
curl http://localhost:8080/health
```

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "database": "up",
    "environment": "development",
    "version": "dev"
  },
  "error": null
}
```

## وضعیت پیاده‌سازی

- **Phase 1** — زیرساخت بک‌اند و اسکلت فرانت‌اند ✅
- **Phase 2** — دیتابیس، مدل‌ها، هسته مالی و تست‌های Integrity ✅
- **Phase 3** — احراز هویت (Session Cookie + CSRF + Rate Limit) ✅
- **Phase 4** — APIهای CRUD مالی + صفحات فرانت‌اند (حساب‌ها، دسته‌ها، هزینه‌ها، فروش، انتقال، تراکنش‌ها، نماینده‌ها، یادآوری‌ها) ✅
- **Phase 5** — داشبورد واقعی + نمودار روند + گزارش‌ها + خروجی CSV ✅
- **Phase 6** — پشتیبان‌گیری (VACUUM INTO با Retention) ✅

مدل مالی کامل در [`docs/financial-model.md`](docs/financial-model.md) مستند است.

## پشتیبان‌گیری

```bash
cd backend
go run ./cmd/backup -db ./data/accounting.db -out ./backups -keep 14
```

اسنپ‌شات سازگار حتی حین فعالیت برنامه می‌گیرد (`VACUUM INTO`)، صحت فایل را با
`integrity_check` بررسی و نسخه‌های قدیمی را مطابق `-keep` حذف می‌کند. برای اجرای
خودکار یک systemd timer روزانه کافی است.

## نکته درباره Migrationها

Migrationها به صورت SQL نسخه‌دار در `backend/migrations/` نگهداری می‌شوند
(الگوی نام‌گذاری سازگار با golang-migrate: `NNNNNN_name.up/down.sql`) و هنگام
اجرای برنامه به‌صورت اتمیک اعمال می‌شوند.
