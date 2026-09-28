"use client";

import { useEffect } from "react";

export function EnsureGuest({ slug, host }: { slug?: string; host?: string }) {
  useEffect(() => {
    const body: { slug?: string; host?: string } = {};
    if (slug) body.slug = slug;
    if (host) body.host = host;
    void fetch("/api/auth/guest", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
  }, [slug, host]);
  return null;
}
