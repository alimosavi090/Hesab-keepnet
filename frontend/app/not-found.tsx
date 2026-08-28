import Link from "next/link";

export default function NotFound() {
  return (
    <main className="bg-background text-foreground flex min-h-dvh flex-col items-center justify-center gap-4 px-6 text-center">
      <p className="text-6xl font-black">۴۰۴</p>
      <p className="text-lg font-semibold">صفحه‌ای که دنبالش بودید پیدا نشد</p>
      <p className="text-caption max-w-sm">
        ممکن است آدرس تغییر کرده باشد یا این بخش هنوز پیاده‌سازی نشده باشد.
      </p>
      <Link
        href="/dashboard"
        className="bg-primary text-primary-foreground mt-2 rounded-lg px-4 py-2 text-sm font-semibold transition-opacity hover:opacity-90"
      >
        بازگشت به داشبورد
      </Link>
    </main>
  );
}
