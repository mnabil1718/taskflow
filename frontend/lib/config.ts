function requirePublicEnv(value: string | undefined, name: string): string {
  if (!value) throw new Error(`Missing required env var: ${name}`);
  return value;
}

export const config = {
  api: {
    // Must use literal access so Next.js can statically inline NEXT_PUBLIC_ vars
    baseUrl: requirePublicEnv(
      process.env.NEXT_PUBLIC_API_URL,
      "NEXT_PUBLIC_API_URL"
    ),
  },
} as const;
