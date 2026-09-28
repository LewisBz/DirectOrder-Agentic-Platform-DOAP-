import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { staffMeRequest } from "@/lib/api";
import { LogoutButton } from "./logout-button";

const roleLabel: Record<string, string> = {
  owner: "dueño",
  cashier: "caja",
  kitchen: "cocina",
};

type Me = {
  name: string;
  role: string;
  tenant_name: string;
  tenant_slug: string;
};

export default async function HelloPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const jar = await cookies();
  const access = jar.get("doap_access")?.value;
  const host = jar.get("doap_host")?.value;
  if (!access || !host) {
    redirect(`/t/${slug}/login`);
  }
  const req = staffMeRequest(access, host);
  const res = await fetch(req.url, req);
  if (!res.ok) {
    redirect(`/t/${slug}/login`);
  }
  const me: Me = await res.json();
  return (
    <main className="mx-auto max-w-xl p-8">
      <h1 className="text-2xl font-semibold">Hola, {me.name}</h1>
      <p className="mt-4 text-zinc-700">
        Rol <strong>{roleLabel[me.role] ?? me.role}</strong> · comercio{" "}
        <strong>{me.tenant_name}</strong> ({me.tenant_slug})
      </p>
      <LogoutButton slug={slug} />
      {me.role === "owner" ? (
        <p className="mt-4">
          <a className="underline" href={`/t/${slug}/staff`}>
            Administrar equipo
          </a>
        </p>
      ) : null}
    </main>
  );
}
