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

export default async function TenantSlugPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const res = await fetch(`${apiURL}/v1/tenants/current/by-slug/${slug}`, {
    cache: "no-store",
  });
  if (!res.ok) {
    return (
      <main className="mx-auto max-w-xl p-8">
        <h1 className="text-2xl font-semibold">Comercio no encontrado</h1>
        <p className="mt-2">Slug {slug}</p>
      </main>
    );
  }
  const tenant: TenantPublic = await res.json();
  return (
    <main className="mx-auto max-w-xl p-8">
      <h1 className="text-2xl font-semibold">{tenant.name}</h1>
      <p className="mt-2 text-zinc-600">
        {tenant.slug} · {tenant.host} · {tenant.currency} · {tenant.tax_name} · {tenant.timezone}
      </p>
    </main>
  );
}
