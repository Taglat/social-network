import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/lib/auth";
import { SocketProvider } from "@/lib/socket";
import AppShell from "@/components/AppShell";

export const metadata: Metadata = {
  title: "Social Network",
  description: "A Facebook-like social network",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          <SocketProvider>
            <AppShell>{children}</AppShell>
          </SocketProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
