import { headers } from "next/headers";
import { apiBaseURL } from "@/lib/api";

const apiURL = apiBaseURL();

type TenantPublic = {
  name: string;
  slug: string;
  host: string;
  currency: string;
  tax_name: string;
  timezone: string;
};

async function loadByHost(host: string): Promise<TenantPublic | null> {
  const res = await fetch(`${apiURL}/v1/tenants/current`, {
    headers: { "X-Forwarded-Host": host },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

export default async function HomePage() {
  const h = await headers();
  const host = h.get("host") ?? "localhost";
  const tenant = await loadByHost(host);

  return (
    <main className="mx-auto max-w-xl p-8">
      <h1 className="text-2xl font-semibold">DOAP</h1>
      {tenant ? (
        <p className="mt-4">
          Comercio: <strong>{tenant.name}</strong> ({tenant.slug}) · {tenant.currency} · {tenant.tax_name} · {tenant.timezone}
        </p>
      ) : (
        <p className="mt-4 text-zinc-600">
          Host <code>{host}</code> no coincide con un comercio. Prueba{" "}
          <a className="underline" href="/t/demo-a">
            /t/demo-a
          </a>
          .
        </p>
      )}
    </main>
  );
}
