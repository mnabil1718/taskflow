function required(key: string): string {
  const value = process.env[key];
  if (!value) throw new Error(`Missing required env var: ${key}`);
  return value;
}

export const config = {
  api: {
    baseUrl: required("NEXT_PUBLIC_API_URL"),
  },
} as const;
