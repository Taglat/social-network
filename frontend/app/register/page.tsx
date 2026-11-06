"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import type { User } from "@/lib/types";

export default function RegisterPage() {
  const { setUser } = useAuth();
  const router = useRouter();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const submit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    const form = new FormData(e.currentTarget);
    // The checkbox value needs normalising to the string the backend expects.
    form.set("isPublic", form.get("isPublic") ? "true" : "false");
    try {
      const user = await api.form<User>("/api/register", form);
      setUser(user);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="center-page">
      <div className="card auth-card">
        <h2 style={{ marginTop: 0 }}>Create account</h2>
        <form onSubmit={submit}>
          <div className="field">
            <label>Email *</label>
            <input name="email" type="email" required />
          </div>
          <div className="field">
            <label>Password * (min 6 characters)</label>
            <input name="password" type="password" minLength={6} required />
          </div>
          <div className="row">
            <div className="field" style={{ flex: 1 }}>
              <label>First name *</label>
              <input name="firstName" required />
            </div>
            <div className="field" style={{ flex: 1 }}>
              <label>Last name *</label>
              <input name="lastName" required />
            </div>
          </div>
          <div className="field">
            <label>Date of birth *</label>
            <input name="dateOfBirth" type="date" required />
          </div>
          <div className="field">
            <label>Avatar (optional)</label>
            <input name="avatar" type="file" accept="image/jpeg,image/png,image/gif" />
          </div>
          <div className="field">
            <label>Nickname (optional)</label>
            <input name="nickname" />
          </div>
          <div className="field">
            <label>About me (optional)</label>
            <textarea name="aboutMe" />
          </div>
          <div className="field row" style={{ gap: 8 }}>
            <input
              name="isPublic"
              type="checkbox"
              style={{ width: "auto" }}
              id="isPublic"
            />
            <label htmlFor="isPublic" style={{ margin: 0 }}>
              Make my profile public
            </label>
          </div>
          {error && <div className="error">{error}</div>}
          <button type="submit" disabled={loading} style={{ width: "100%" }}>
            {loading ? "Creating…" : "Register"}
          </button>
        </form>
        <p className="muted" style={{ marginTop: 14 }}>
          Already have an account? <Link href="/login">Log in</Link>
        </p>
      </div>
    </div>
  );
}
