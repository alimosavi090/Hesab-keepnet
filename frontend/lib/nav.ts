import {
  IconArrowsExchange,
  IconArrowsLeftRight,
  IconBell,
  IconBuildingBank,
  IconChartBar,
  IconLayoutDashboard,
  IconNotes,
  IconReceipt,
  IconSettings,
  IconShoppingCart,
  IconUsersGroup,
  type TablerIcon,
} from "@tabler/icons-react";

export type NavItem = {
  href: string;
  label: string;
  icon: TablerIcon;
};

export const MAIN_NAV: NavItem[] = [
  { href: "/dashboard", label: "داشبورد", icon: IconLayoutDashboard },
  { href: "/sales", label: "فروش", icon: IconShoppingCart },
  { href: "/expenses", label: "هزینه‌ها", icon: IconReceipt },
  { href: "/representatives", label: "نماینده‌ها", icon: IconUsersGroup },
  { href: "/bank-accounts", label: "حساب‌ها", icon: IconBuildingBank },
  { href: "/transfers", label: "انتقال‌ها", icon: IconArrowsExchange },
  { href: "/transactions", label: "تراکنش‌ها", icon: IconArrowsLeftRight },
  { href: "/reports", label: "گزارش‌ها", icon: IconChartBar },
  { href: "/journal", label: "دفتر روزانه", icon: IconNotes },
  { href: "/reminders", label: "یادآوری‌ها", icon: IconBell },
  { href: "/settings", label: "تنظیمات", icon: IconSettings },
];

export const MOBILE_NAV_HREFS = [
  "/dashboard",
  "/sales",
  "/expenses",
  "/transactions",
];

export function mobileNavItems(): NavItem[] {
  return MAIN_NAV.filter((item) => MOBILE_NAV_HREFS.includes(item.href));
}

export function secondaryNavItems(): NavItem[] {
  return MAIN_NAV.filter((item) => !MOBILE_NAV_HREFS.includes(item.href));
}

export function isActivePath(pathname: string, href: string): boolean {
  if (href === "/dashboard") return pathname === href;
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function pageTitle(pathname: string): string {
  const match = [...MAIN_NAV]
    .sort((a, b) => b.href.length - a.href.length)
    .find((item) => isActivePath(pathname, item.href));
  return match?.label ?? "حساب‌کیپ";
}
