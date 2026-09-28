"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";

type Staff = {
  id: string;
  email: string;
  name: string;
  role: string;
  active: boolean;
};

const roleLabel: Record<string, string> = {
  owner: "dueño",
  cashier: "caja",
  kitchen: "cocina",
};

export default function StaffPage() {
  const { slug } = useParams<{ slug: string }>();
  const router = useRouter();
  const [items, setItems] = useState<Staff[]>([]);
  const [error, setError] = useState("");
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("cashier");
  const [editId, setEditId] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [editRole, setEditRole] = useState("cashier");
  const [editPassword, setEditPassword] = useState("");

  async function load() {
    const res = await fetch("/api/staff", { cache: "no-store" });
    if (res.status === 401) {
      router.push(`/t/${slug}/login`);
      return;
    }
    if (res.status === 403) {
      setError("Solo el dueño administra el equipo.");
      setItems([]);
      return;
    }
    if (!res.ok) {
      setError("No se pudo listar el equipo.");
      return;
    }
    setError("");
    setItems(await res.json());
  }

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [slug]);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    const res = await fetch("/api/staff", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, name, password, role }),
    });
    if (!res.ok) {
      setError("No se pudo crear. Revisa correo único, rol caja/cocina y clave ≥ 8.");
      return;
    }
    setEmail("");
    setName("");
    setPassword("");
    await load();
  }

  async function onPatch(id: string) {
    const body: Record<string, string> = { name: editName, role: editRole };
    if (editPassword) body.password = editPassword;
    const res = await fetch(`/api/staff/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      setError("No se pudo editar.");
      return;
    }
    setEditId(null);
    setEditPassword("");
    await load();
  }

  async function toggle(id: string, active: boolean) {
    const path = active ? "deactivate" : "reactivate";
    const res = await fetch(`/api/staff/${id}/${path}`, { method: "POST" });
    if (!res.ok) {
      setError("No se pudo cambiar el estado (el dueño no se desactiva).");
      return;
    }
    await load();
  }

  return (
    <main className="mx-auto max-w-xl p-8">
      <h1 className="text-2xl font-semibold">Equipo</h1>
      <p className="mt-1 text-sm text-zinc-600">
        <a className="underline" href={`/t/${slug}/hello`}>
          Saludo
        </a>
      </p>
      {error ? <p className="mt-3 text-sm text-red-700">{error}</p> : null}

      <ul className="mt-6 space-y-3">
        {items.map((s) => (
          <li key={s.id} className="rounded border border-zinc-200 p-3">
            {editId === s.id ? (
              <div className="flex flex-col gap-2">
                <input
                  className="rounded border px-2 py-1"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                />
                <select
                  className="rounded border px-2 py-1"
                  value={editRole}
                  onChange={(e) => setEditRole(e.target.value)}
                >
                  <option value="cashier">caja</option>
                  <option value="kitchen">cocina</option>
                </select>
                <input
                  className="rounded border px-2 py-1"
                  type="password"
                  placeholder="Nueva clave (opcional)"
                  value={editPassword}
                  onChange={(e) => setEditPassword(e.target.value)}
                />
                <button className="rounded bg-zinc-900 px-3 py-1 text-white" type="button" onClick={() => onPatch(s.id)}>
                  Guardar
                </button>
              </div>
            ) : (
              <>
                <p>
                  <strong>{s.name}</strong> · {s.email} · {roleLabel[s.role] ?? s.role} ·{" "}
                  {s.active ? "activo" : "inactivo"}
                </p>
                {s.role !== "owner" ? (
                  <div className="mt-2 flex gap-2 text-sm">
                    <button
                      className="underline"
                      type="button"
                      onClick={() => {
                        setEditId(s.id);
                        setEditName(s.name);
                        setEditRole(s.role);
                      }}
                    >
                      Editar
                    </button>
                    <button className="underline" type="button" onClick={() => toggle(s.id, s.active)}>
                      {s.active ? "Desactivar" : "Reactivar"}
                    </button>
                  </div>
                ) : null}
              </>
            )}
          </li>
        ))}
      </ul>

      <form className="mt-8 flex flex-col gap-2" onSubmit={onCreate}>
        <h2 className="font-medium">Alta caja o cocina</h2>
        <input className="rounded border px-2 py-1" placeholder="Nombre" value={name} onChange={(e) => setName(e.target.value)} required />
        <input className="rounded border px-2 py-1" type="email" placeholder="Correo" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <input className="rounded border px-2 py-1" type="password" placeholder="Contraseña (≥8)" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
        <select className="rounded border px-2 py-1" value={role} onChange={(e) => setRole(e.target.value)}>
          <option value="cashier">caja</option>
          <option value="kitchen">cocina</option>
        </select>
        <button className="rounded bg-zinc-900 px-4 py-2 text-white" type="submit">
          Crear
        </button>
      </form>
    </main>
  );
}
