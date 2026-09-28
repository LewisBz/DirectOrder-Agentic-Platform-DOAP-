export function apiBaseURL(): string {
  return (
    process.env.API_INTERNAL_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    "http://localhost:8080"
  );
}

export function staffMeRequest(accessToken: string, host: string): RequestInit & { url: string } {
  return {
    url: `${apiBaseURL()}/v1/auth/me`,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "X-Forwarded-Host": host,
    },
    cache: "no-store",
  };
}
