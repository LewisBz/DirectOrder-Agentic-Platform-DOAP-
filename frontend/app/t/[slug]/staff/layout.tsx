import { cookies } from "next/headers";
import { redirect } from "next/navigation";

export default async function StaffLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const jar = await cookies();
  if (!jar.get("doap_access")?.value) {
    redirect(`/t/${slug}/login`);
  }
  return children;
}
