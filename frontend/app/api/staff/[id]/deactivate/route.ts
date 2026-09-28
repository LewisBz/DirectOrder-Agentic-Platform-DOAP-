import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiBaseURL } from "@/lib/api";

export async function POST(_req: Request, ctx: { params: Promise<{ id: string }> }) {
  const jar = await cookies();
  const access = jar.get("doap_access")?.value;
  const host = jar.get("doap_host")?.value;
  if (!access || !host) return NextResponse.json({ error: "invalid_credentials" }, { status: 401 });
  const { id } = await ctx.params;
  const res = await fetch(`${apiBaseURL()}/v1/staff/${id}/deactivate`, {
    method: "POST",
    headers: { Authorization: `Bearer ${access}`, "X-Forwarded-Host": host },
    cache: "no-store",
  });
  if (res.status === 204) return new NextResponse(null, { status: 204 });
  const body = await res.json().catch(() => ({}));
  return NextResponse.json(body, { status: res.status });
}
