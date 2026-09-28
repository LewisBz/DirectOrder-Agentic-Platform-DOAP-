import { NextResponse } from "next/server";
import { apiBaseURL } from "@/lib/api";

const accessMaxAge = 15 * 60;
const refreshMaxAge = 7 * 24 * 60 * 60;

export async function POST(req: Request) {
  const body = (await req.json()) as { email?: string; password?: string; slug?: string };
  const slug = body.slug ?? "";
  const api = apiBaseURL();
  const tenantRes = await fetch(`${api}/v1/tenants/current/by-slug/${encodeURIComponent(slug)}`, {
    cache: "no-store",
  });
  if (!tenantRes.ok) {
    return NextResponse.json({ error: "tenant_not_found" }, { status: 404 });
  }
  const tenant = (await tenantRes.json()) as { host: string };
  const loginRes = await fetch(`${api}/v1/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Forwarded-Host": tenant.host,
    },
    body: JSON.stringify({ email: body.email, password: body.password }),
    cache: "no-store",
  });
  const payload = await loginRes.json().catch(() => ({}));
  if (!loginRes.ok) {
    return NextResponse.json(payload, { status: loginRes.status });
  }
  const res = NextResponse.json({ ok: true });
  const secure = process.env.NODE_ENV === "production";
  res.cookies.set("doap_access", payload.access_token, {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: accessMaxAge,
    secure,
  });
  res.cookies.set("doap_refresh", payload.refresh_token, {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: refreshMaxAge,
    secure,
  });
  res.cookies.set("doap_host", tenant.host, {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: refreshMaxAge,
    secure,
  });
  return res;
}
