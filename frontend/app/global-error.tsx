"use client";

import Link from "next/link";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="fa" dir="rtl">
      <body
        style={{
          minHeight: "100dvh",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          gap: "1.25rem",
          fontFamily: "Vazirmatn, system-ui, sans-serif",
          background: "#16181d",
          color: "#e9eaee",
          padding: "1.5rem",
        }}
      >
        <p style={{ fontSize: "1.125rem", fontWeight: 700 }}>
          خطای بحرانی در اجرای برنامه
        </p>
        <p style={{ fontSize: "0.875rem", opacity: 0.7 }}>
          متأسفانه برنامه قادر به ادامه اجرا نیست. لطفاً صفحه را دوباره بارگذاری
          کنید.
        </p>
        <div style={{ display: "flex", gap: "0.75rem" }}>
          <button
            onClick={reset}
            style={{
              padding: "0.5rem 1rem",
              borderRadius: "0.5rem",
              background: "#e9eaee",
              color: "#16181d",
              fontWeight: 600,
            }}
          >
            تلاش مجدد
          </button>
          <Link href="/dashboard">
            <span
              style={{
                display: "inline-block",
                padding: "0.5rem 1rem",
                borderRadius: "0.5rem",
                border: "1px solid rgba(233,234,238,0.3)",
              }}
            >
              داشبورد
            </span>
          </Link>
        </div>
        {process.env.NODE_ENV === "development" ? (
          <pre dir="ltr" style={{ fontSize: "0.75rem", opacity: 0.6, maxWidth: "32rem", overflow: "auto" }}>
            {error.message}
          </pre>
        ) : null}
      </body>
    </html>
  );
}
