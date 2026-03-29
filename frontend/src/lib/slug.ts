/**
 * Converts an entity name to a URL-safe slug.
 * Example: "Feed API" → "feed-api"
 */
export function slug(name: string): string {
  return name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
}
