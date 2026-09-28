import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiBaseURL } from "@/lib/api";

async function creds() {
  const jar = await cookies();
  const access = jar.get("doap_access")?.value;
  const host = jar.get("doap_host")?.value;
  if (!access || !host) return null;
  return { access, host };
}

export async function PATCH(req: Request, ctx: { params: Promise<{ id: string }> }) {
  const c = await creds();
  if (!c) return NextResponse.json({ error: "invalid_credentials" }, { status: 401 });
  const { id } = await ctx.params;
  const payload = await req.json();
  const res = await fetch(`${apiBaseURL()}/v1/staff/${id}`, {
    method: "PATCH",
    headers: {
      Authorization: `Bearer ${c.access}`,
      "X-Forwarded-Host": c.host,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
    cache: "no-store",
  });
  const body = await res.json().catch(() => ({}));
  return NextResponse.json(body, { status: res.status });
}
