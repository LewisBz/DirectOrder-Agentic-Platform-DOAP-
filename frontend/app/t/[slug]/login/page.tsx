"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";

export default function LoginPage() {
  const router = useRouter();
  const { slug } = useParams<{ slug: string }>();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setPending(true);
    const res = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password, slug }),
    });
    setPending(false);
    if (!res.ok) {
      setError("No se pudo entrar. Revisa correo y contraseña.");
      return;
    }
    router.push(`/t/${slug}/hello`);
    router.refresh();
  }

  return (
    <main className="mx-auto max-w-sm p-8">
      <h1 className="text-2xl font-semibold">Acceso del comercio</h1>
      <form className="mt-6 flex flex-col gap-3" onSubmit={onSubmit}>
        <label className="text-sm">
          Correo
          <input
            className="mt-1 w-full rounded border border-zinc-300 px-3 py-2"
            type="email"
            value={email}
            onChange={(ev) => setEmail(ev.target.value)}
            required
          />
        </label>
        <label className="text-sm">
          Contraseña
          <input
            className="mt-1 w-full rounded border border-zinc-300 px-3 py-2"
            type="password"
            value={password}
            onChange={(ev) => setPassword(ev.target.value)}
            required
          />
        </label>
        {error ? <p className="text-sm text-red-700">{error}</p> : null}
        <button
          className="rounded bg-zinc-900 px-4 py-2 text-white disabled:opacity-50"
          disabled={pending}
          type="submit"
        >
          Entrar
        </button>
      </form>
    </main>
  );
}
