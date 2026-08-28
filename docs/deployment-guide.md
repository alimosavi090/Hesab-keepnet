# 🚀 راهنمای کامل دیپلوی روی سرور (VPS Ubuntu)

معماری نهایی:

```
Internet → Nginx (443, HTTPS) ─┬→ /api/* و /health  →  backend :8080  ← SQLite در /opt/hesab/data
                               └→ /*                 →  next.js  :3000
```

هر دو سرویس با `systemd` مدیریت می‌شوند. مقادیر نمونه: دامنه `hesab.example.com`، کاربر `deploy`.

---

## مرحله ۰ — پیش‌نیازها روی سرور

سرور Ubuntu 22/24 با حداقل ۱GB RAM (اگر کمتر، swap بسازید).

```bash
adduser deploy && usermod -aG sudo deploy
su - deploy

# فایروال
sudo ufw allow OpenSSH && sudo ufw enable

# Go و Node و Nginx
sudo apt update && sudo apt install -y nginx git curl
sudo snap install go --classic          # یا از go.dev/dl باینری نصب کنید
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt install -y nodejs
```

---

## مرحله ۱ — دریافت کد

```bash
sudo mkdir -p /opt/hesab && sudo chown deploy:deploy /opt/hesab
cd /opt/hesab
git clone git@github.com:<USERNAME>/hesab-keepnet.git app
cd app
```

> برای اینکه clone با SSH روی سرور انجام شود، یک SSH key مخصوص سرور بسازید (`ssh-keygen`) و Deploy Key فقط-خواندنی آن را به گیتهاب اضافه کنید: `Settings → Deploy keys → Add` با تیک **Read-only**.

---

## مرحله ۲ — ساخت بک‌اند

```bash
cd /opt/hesab/app/backend
go build -o bin/hesab-backend ./cmd/server
```

### فایل محیطی بک‌اند: `/opt/hesab/app/backend/.env`

```env
APP_ENV=production
APP_PORT=8080
DATABASE_PATH=/opt/hesab/data/accounting.db
BACKUP_DIR=/opt/hesab/backups
BACKUP_INTERVAL_HOURS=24

# دامنه فرانت (بدون / انتهایی، چندتا با کاما)
CORS_ORIGIN=https://hesab.example.com
COOKIE_SECURE=true

# حداقل ۱۶ کاراکتر؛ با: openssl rand -hex 32
SESSION_SECRET=<REPLACE_WITH_LONG_RANDOM>

# ⚠️ این دو فقط موقع seed اولیه لازم می‌شوند (مرحله ۴)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=<REPLACE_STRONG>
```

⚠️ `chmod 600 .env` — این فایل پسورد دارد.

---

## مرحله ۳ — ساخت فرانت‌اند

نکته معماری: اگر Nginx همان دامنه را به هر دو سرویس پروکسی کند (که پایین همین کار را می‌کنیم)، آدرس API نسبت-به‌دامنه می‌شود و فرانت باید این‌طور ساخته شود:

```bash
cd /opt/hesab/app/frontend

cat > .env.local <<'EOF'
NEXT_PUBLIC_API_URL=
EOF
# مقدار خالی یعنی «همان دامنه فعلی مرورگر» — امن‌ترین گزینه

npm install   # یا npm ci اگر lockfile در ریپوست
npm run build
```

---

## مرحله ۴ — اولین اجرا و seed ادمین ⚠️

در حالت `production` سیستم کاربرseed نمی‌کند؛ پس یک بار در حالت development اجرا کنید تا ادمین ساخته شود:

```bash
cd /opt/hesab/app/backend
mkdir -p /opt/hesab/data

# اجرای موقت برای ساخت دیتابیس + کاربر admin:
APP_ENV=development \
ADMIN_USERNAME=admin ADMIN_PASSWORD=<REPLACE_STRONG> \
./bin/hesab-backend &
sleep 3 && kill %1

# از این به بعد APP_ENV=production است و password_hash ساخته‌شده در DB می‌ماند ✅
```

تست سریع سلامت:

```bash
curl http://localhost:8080/health
# {"status":"ok","database":"up", ...}
```

---

## مرحله ۵ — systemd (اجرای دائمی)

### `/etc/systemd/system/hesab-backend.service`

```ini
[Unit]
Description=Hesab-keepnet backend (Go)
After=network.target

[Service]
User=deploy
WorkingDirectory=/opt/hesab/app/backend
EnvironmentFile=/opt/hesab/app/backend/.env
ExecStart=/opt/hesab/app/backend/bin/hesab-backend
Restart=always
RestartSec=5
# Hardening
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=/opt/hesab/data /opt/hesab/backups
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

### `/etc/systemd/system/hesab-frontend.service`

```ini
[Unit]
Description=Hesab-keepnet frontend (Next.js)
After=network.target hesab-backend.service

[Service]
User=deploy
WorkingDirectory=/opt/hesab/app/frontend
Environment=NODE_ENV=production
ExecStart=/usr/bin/npm run start -- -p 3000
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

فعال‌سازی:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now hesab-backend hesab-frontend
sudo systemctl status hesab-backend   # باید active (running) باشد
```

*نکته: backend سرویس داده خودکار پشتیبان‌گیری را هم بالا می‌آورد (خودش هندل می‌کند).*

---

## مرحله ۶ — Nginx + HTTPS

### `/etc/nginx/sites-available/hesab`

```nginx
server {
    listen 80;
    server_name hesab.example.com;

    # بکاپ‌ها/دیتابیس از بیرون اصلاً سرو نمی‌شوند
    location ~ ^/(api|health) {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto https;
    }

    location / {
        proxy_pass         http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection "upgrade";
        proxy_set_header   Host $host;
        proxy_set_header   X-Forwarded-Proto https;
    }

    client_max_body_size 20m;
}
```

```bash
sudo ln -s /etc/nginx/sites-available/hesab /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# HTTPS رایگان با Let's Encrypt:
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d hesab.example.com
# تیک redirect خودکار HTTP→HTTPS بزنید. تمدید خودکار نصب شده است.
```

🎉 الان `https://hesab.example.com` لاگین صفحه را نشان می‌دهد.

---

## مرحله ۷ — پشتیبان و بازگردانی

```bash
# فهرست پشتیبان‌های خودکار از داخل اپ: تنظیمات ← کارت پشتیبان‌گیری
ls -lh /opt/hesab/backups/

# کپی دور از سرور (مهم!) — روی لپ‌تاپ خودتان مثلاً:
scp deploy@hesab.example.com:/opt/hesab/backups/auto-*.db ~/backups/
```

بازگردانی (restore):

```bash
sudo systemctl stop hesab-backend
cp auto-20260827_033000.db /opt/hesab/data/accounting.db
sudo systemctl start hesab-backend
```

⚠️ چون بکاپ و دیسک یکی هستند، یک cron ماهانه هم اجباری کنید که فایل آخرین بکاپ را جایی خارج از سرور ببرد:

```bash
crontab -e
# 15 4 * * * rsync -az /opt/hesab/backups/ backupbox:/mnt/hesab-backups/
```

---

## مرحله ۸ — نگهداری روزمره

| کار | دستور |
|---|---|
| انتشار نسخه جدید | `cd /opt/hesab/app && git pull && cd backend && go build -o bin/hesab-backend ./cmd/server && cd ../frontend && npm ci \|\| npm install && npm run build` سپس `sudo systemctl restart hesab-backend hesab-frontend` |
| لاگ زنده | `journalctl -u hesab-backend -f` |
| وضعیت | `systemctl status hesab-*` |
| تغییر رمز SESSION_SECRET | `.env` را عوض کنید و restart (لاگین همه خالی می‌شود) |

---

## چک‌لیست نهایی Production

- [ ] `https://` فعال و HTTP ریدایرکت می‌شود
- [ ] `COOKIE_SECURE=true` و `SESSION_SECRET` تصادفی ≥۱۶ کاراکتر
- [ ] پسورد ادمین قوی و ذخیره در Password Manager
- [ ] `ufw`: فقط 22، 80، 443 باز — پورت 8080/3000 **از بیرون بسته**
- [ ] پوشه `/opt/hesab/app/backend/.env` با `chmod 600`
- [ ] یک پشتیبان دانلود و بیرون از سرور کپی شده
- [ ] `unattended-upgrades` فعال (`sudo dpkg-reconfigure -plow unattended-upgrades`)
- [ ] بعد از چند روز: حجم log ها و مصرف RAM بررسی شود

مشکل‌یابی سریع:
- **502 روی /api** → بک‌اند down است: `systemctl status hesab-backend` و لاگش را ببینید
- **صفحه سفید فرانت** → `NEXT_PUBLIC_API_URL` احتمالاً localhost مانده؛ rebuild لازم است
- **کراس‌داومین خطا در کنسول** → مطمئن شوید CORS_ORIGIN دقیقاً origin فرانت را دارد، یا کل مسیر same-origin باشد (این راهنما same-origin است)
