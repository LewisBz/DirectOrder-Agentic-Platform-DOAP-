import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiBaseURL } from "@/lib/api";

export async function POST() {
  const jar = await cookies();
  const refresh = jar.get("doap_refresh")?.value;
  const host = jar.get("doap_host")?.value;
  if (!refresh || !host) {
    return NextResponse.json({ error: "invalid_credentials" }, { status: 401 });
  }
  const apiRes = await fetch(`${apiBaseURL()}/v1/auth/refresh`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Forwarded-Host": host,
    },
    body: JSON.stringify({ refresh_token: refresh }),
    cache: "no-store",
  });
  const payload = await apiRes.json().catch(() => ({}));
  if (!apiRes.ok) {
    const res = NextResponse.json(payload, { status: apiRes.status });
    res.cookies.delete("doap_access");
    res.cookies.delete("doap_refresh");
    return res;
  }
  const res = NextResponse.json({ ok: true });
  const secure = process.env.NODE_ENV === "production";
  res.cookies.set("doap_access", payload.access_token, {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: 15 * 60,
    secure,
  });
  res.cookies.set("doap_refresh", payload.refresh_token, {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: 7 * 24 * 60 * 60,
    secure,
  });
  return res;
}
