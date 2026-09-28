import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiBaseURL } from "@/lib/api";

export async function POST() {
  const jar = await cookies();
  const access = jar.get("doap_access")?.value;
  const host = jar.get("doap_host")?.value;
  if (access && host) {
    await fetch(`${apiBaseURL()}/v1/auth/logout`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${access}`,
        "X-Forwarded-Host": host,
      },
      cache: "no-store",
    });
  }
  const res = NextResponse.json({ ok: true });
  res.cookies.delete("doap_access");
  res.cookies.delete("doap_refresh");
  res.cookies.delete("doap_host");
  return res;
}
