"use client";

import { useEffect, useState } from "react";

const apiURL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default function HomePage() {
  const [health, setHealth] = useState<string>("checking…");

  useEffect(() => {
    fetch(`${apiURL}/healthz`)
      .then((r) => r.json())
      .then((body: { status?: string }) => setHealth(body.status ?? "unknown"))
      .catch(() => setHealth("unreachable"));
  }, []);

  return (
    <main className="mx-auto max-w-xl p-8">
      <h1 className="text-2xl font-semibold">DOAP</h1>
      <p className="mt-2 text-zinc-600">Vitrina mínima. API: {apiURL}</p>
      <p className="mt-4">
        Señal de vida: <span className="font-medium">{health}</span>
      </p>
    </main>
  );
}
