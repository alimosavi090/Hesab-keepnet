# 📤 راهنمای کامل انتشار پروژه روی GitHub

این راهنما برای ریپوی **Hesab-keepnet** (بک‌اند Go + فرانت‌اند Next.js) نوشته شده است. از ابتدای صفر تا یک ریپوی تمیز با CI.

---

## مرحله ۱ — آماده‌سازی امنیتی (مهم‌ترین مرحله!)

قبل از هر push، مطمئن شو هیچ داده یا رمزی داخل ریپو نیست:

```bash
# این فایل‌ها نباید track شوند:
cat .gitignore
```

`.gitignore` این پروژه از قبل شامل موارد حیاتی است:
- `.env*` — تمام کلیدهای محیطی (SESSION_SECRET و…)
- `data/` ، `*.db` — دیتابیس SQLite و پشتیبان‌ها ⚠️ **هرگز دیتابیس واقعی به گیتهاب نروید**
- `node_modules/`, `.next/`, `backend/bin/`

### بررسی دستی قبل از اولین push

```bash
# آیا جایی رمز هاردکد نشده؟
grep -rn "SESSION_SECRET\|password\s*=\|api_key" --include="*.go" --include="*.ts" --include="*.tsx" --include="*.env*" .
grep -rniE "(secret|token|pass)\s*[:=]\s*[\"'][^\"']{8}" backend/internal frontend/app frontend/lib | grep -v test
```

اگر فایلی قبلاً با داده حساس commit شده و تاریخچه هنوز منتشر نشده، بهترین کار بازنویسی تاریخچه است (`git filter-repo`) — اما چون هنوز remote ندارید، احتمالاً لازم نیست.

---

## مرحله ۲ — ساخت ریپو روی GitHub

1. وارد [github.com/new](https://github.com/new) شوید
2. نام ریپو: `hesab-keepnet`
3. **Private** انتخاب کنید ✅ — این نرم‌افزار مالی شخصی شماست؛ عمومی کردن هیچ مزایایی ندارد
4. تیک‌های README/.gitignore/license را **خالی** بگذارید (الان داریم)
5. Create repository

---

## مرحله ۳ — اتصال و اولین push

```bash
cd ~/Projects/Hesab-keepnet

# بررسی وضعیت فعلی
git status
git log --oneline -5

# اتصال ریموت (username خودتان را جایگزین کنید)
git remote add origin git@github.com:<USERNAME>/hesab-keepnet.git

# اگر SSH تنظیم نکرده‌اید:
#   ssh-keygen -t ed25519 -C "you@email.com"
#   cat ~/.ssh/id_ed25519.pub   ← در Settings → SSH keys گیتهاب ثبت کنید

git branch -M main
git push -u origin main
```

**بررسی بعد از push:** در صفحه ریپو باید دیرکتوری‌های `backend/`, `frontend/`, `docs/` دیده شود ولی نه `data/` و نه هیچ `.env`.

---

## مرحله ۴ — توضیحات و متادیتا

بدنه README از قبل موجود است؛ اگر بخش دمو/راه‌اندازی را غنی‌تر خواستید این خطوط را به README اضافه کنید:

```markdown
## Quick Start
cp backend/.env.example backend/.env   # سپس مقادیر را پر کنید
cp frontend/.env.example.local frontend/frontend/.env.local
```

(در همین ریپو دو فایل نمونه `.env.example` بسازید تا کسی مجبور نباشد مستندات را بگردد.)

### Branch protection (پیشنهادی)
در `Settings → Branches → Add rule` برای `main`:
- ☑ Require pull request reviews — فقط اگر تیم دارید؛ تک‌نفره لازم نیست
- ☑ Require status checks to pass (بعد از ساخت CI)

---

## مرحله ۵ — CI با GitHub Actions (اختیاری ولی ارزشمند)

فایل `backend/.github/workflows/ci.yml`؟ — مسیر صحیح در ریشه است: `.github/workflows/ci.yml`

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: backend/go.mod
      - name: Build & Vet & Test
        working-directory: backend
        run: |
          go build ./...
          go vet ./...
          go test ./...

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - name: Install
        working-directory: frontend
        run: npm ci || npm install
      - name: Lint & Build
        working-directory: frontend
        run: |
          npm run lint
          npm run build
```

از این لحظه هر push رنگ سبز/قرمز وضعیت تست‌ها را نشان می‌دهد.

---

## گردش روزمره

```bash
# شروع هر کار جدید
git checkout -b feat/xyz

# کامیت‌های کوچک و معنادار
git add <files>
git commit -m "feat(reports): افزودن مقایسه دوره قبل"

# ادغام در main
git checkout main && git merge --no-ff feat/xyz
git push

# نگه داشتن فورک محلی به‌روز از سرور دیگر:
git pull --rebase origin main
```

قاعده پیام کامیت: `<type>(<scope>): <خلاصه>` — انواع رایج: `feat`, `fix`, `refactor`, `docs`, `chore`.

---

## چک‌لیست نهایی انتشار

- [ ] `data/` و `*.db` در گیتهاب دیده نمی‌شوند
- [ ] هیچ `.env` کامیت نشده؛ فقط فایل‌های `*.example`
- [ ] ریپو Private است
- [ ] SSH key تنظیم شده (نه HTTPS + token خام)
- [ ] CI سبز است
- [ ] README حداقل یک «Quick Start» دارد
