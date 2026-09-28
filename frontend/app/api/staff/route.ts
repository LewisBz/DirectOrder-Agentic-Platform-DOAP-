import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiBaseURL } from "@/lib/api";

async function staffHeaders(): Promise<{ access: string; host: string } | null> {
  const jar = await cookies();
  const access = jar.get("doap_access")?.value;
  const host = jar.get("doap_host")?.value;
  if (!access || !host) return null;
  return { access, host };
}

export async function GET() {
  const creds = await staffHeaders();
  if (!creds) return NextResponse.json({ error: "invalid_credentials" }, { status: 401 });
  const res = await fetch(`${apiBaseURL()}/v1/staff`, {
    headers: { Authorization: `Bearer ${creds.access}`, "X-Forwarded-Host": creds.host },
    cache: "no-store",
  });
  const body = await res.json().catch(() => ({}));
  return NextResponse.json(body, { status: res.status });
}

export async function POST(req: Request) {
  const creds = await staffHeaders();
  if (!creds) return NextResponse.json({ error: "invalid_credentials" }, { status: 401 });
  const payload = await req.json();
  const res = await fetch(`${apiBaseURL()}/v1/staff`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${creds.access}`,
      "X-Forwarded-Host": creds.host,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
    cache: "no-store",
  });
  const body = await res.json().catch(() => ({}));
  return NextResponse.json(body, { status: res.status });
}
