import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { apiBaseURL } from "@/lib/api";

const guestMaxAge = 365 * 24 * 60 * 60;

function hostWithoutPort(host: string): string {
  const i = host.lastIndexOf(":");
  if (i > -1 && !host.includes("]")) {
    return host.slice(0, i);
  }
  return host;
}

export async function POST(req: Request) {
  const jar = await cookies();
  const existing = jar.get("doap_guest")?.value ?? "";
  let slug = "";
  let host = "";
  try {
    const body = (await req.json()) as { slug?: string; host?: string };
    slug = body.slug ?? "";
    host = body.host ? hostWithoutPort(body.host) : "";
  } catch {
    /* empty body */
  }
  const api = apiBaseURL();
  if (slug) {
    const tenantRes = await fetch(`${api}/v1/tenants/current/by-slug/${encodeURIComponent(slug)}`, {
      cache: "no-store",
    });
    if (!tenantRes.ok) {
      return NextResponse.json({ error: "tenant_not_found" }, { status: 404 });
    }
    const tenant = (await tenantRes.json()) as { host: string };
    host = tenant.host;
  }
  if (!host) {
    return NextResponse.json({ error: "tenant_not_found" }, { status: 404 });
  }
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-Forwarded-Host": host,
  };
  if (existing) {
    headers["X-Guest-Token"] = existing;
  }
  const apiRes = await fetch(`${api}/v1/guest/sessions`, {
    method: "POST",
    headers,
    body: JSON.stringify({ channel: "pwa" }),
    cache: "no-store",
  });
  const payload = (await apiRes.json().catch(() => ({}))) as { token?: string };
  if (!apiRes.ok || !payload.token) {
    return NextResponse.json(payload, { status: apiRes.status });
  }
  const res = NextResponse.json({ ok: true, reused: apiRes.status === 200 });
  const secure = process.env.NODE_ENV === "production";
  res.cookies.set("doap_guest", payload.token, {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: guestMaxAge,
    secure,
  });
  return res;
}
