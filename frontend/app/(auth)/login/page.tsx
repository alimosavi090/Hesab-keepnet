"use client";

import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { motion } from "framer-motion";
import { IconLoader2 } from "@tabler/icons-react";
import { authApi } from "@/lib/api";
import { SPRING } from "@/components/shared/motion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setPending(true);
    try {
      await authApi.login(username, password);
      router.replace("/dashboard");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "ورود ناموفق بود.");
    } finally {
      setPending(false);
    }
  }

  return (
    <main className="flex min-h-dvh items-center justify-center px-4">
      {/* Floating aurora orbs */}
      <motion.div
        aria-hidden="true"
        className="animate-float bg-primary/25 pointer-events-none fixed -top-24 left-1/4 -z-10 size-80 rounded-full blur-3xl"
      />
      <motion.div
        aria-hidden="true"
        className="animate-float bg-chart-3/20 pointer-events-none fixed -right-16 bottom-10 -z-10 size-72 rounded-full blur-3xl [animation-delay:-4s]"
      />

      <motion.div
        initial={{ opacity: 0, y: 32, scale: 0.96 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ ...SPRING, delay: 0.05 }}
        className="w-full max-w-sm"
      >
        <Card className="glass lift sheen border-transparent py-8 shadow-2xl ring-foreground/[0.06]">
          <CardHeader className="text-center">
            <div className="mb-4 flex justify-center">
              <span className="brand-tile animate-pulse-glow flex size-14 items-center justify-center rounded-2xl text-2xl font-black text-primary-foreground transition-transform duration-300 hover:scale-105">
                ح
              </span>
            </div>
            <CardTitle className="text-gradient text-2xl font-extrabold">حساب‌کیپ</CardTitle>
            <CardDescription>برای ادامه وارد حساب کاربری شوید</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="username">نام کاربری</Label>
                <Input
                  id="username"
                  dir="ltr"
                  autoComplete="username"
                  className="transition-all duration-300 focus-visible:shadow-[0_0_0_4px_color-mix(in_oklch,var(--primary)_18%,transparent)]"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="password">گذرواژه</Label>
                <Input
                  id="password"
                  type="password"
                  dir="ltr"
                  autoComplete="current-password"
                  className="transition-all duration-300 focus-visible:shadow-[0_0_0_4px_color-mix(in_oklch,var(--primary)_18%,transparent)]"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>

              {error ? (
                <motion.p
                  role="alert"
                  initial={{ opacity: 0, x: -6 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ duration: 0.3 }}
                  className="text-destructive text-sm"
                >
                  {error}
                </motion.p>
              ) : null}

              <Button
                type="submit"
                className="glow-primary h-11 w-full rounded-xl text-base font-bold transition-transform duration-300 hover:scale-[1.02] active:scale-[0.98]"
                disabled={pending}
              >
                {pending ? <IconLoader2 className="size-4 animate-spin" /> : null}
                ورود
              </Button>
            </form>
          </CardContent>
        </Card>
      </motion.div>
    </main>
  );
}
