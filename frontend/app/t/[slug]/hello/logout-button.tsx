"use client";

import { useRouter } from "next/navigation";

export function LogoutButton({ slug }: { slug: string }) {
  const router = useRouter();
  return (
    <button
      className="mt-6 rounded border border-zinc-300 px-4 py-2"
      type="button"
      onClick={async () => {
        await fetch("/api/auth/logout", { method: "POST" });
        router.push(`/t/${slug}/login`);
        router.refresh();
      }}
    >
      Cerrar sesión
    </button>
  );
}
