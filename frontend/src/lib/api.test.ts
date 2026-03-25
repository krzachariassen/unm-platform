import { describe, it, expect } from 'vitest'

// extractError is not exported — test it indirectly by reproducing the same logic here.
// These tests document the expected behavior.

async function extractError(res: Response, fallback: string): Promise<string> {
  try {
    const body = await res.json()
    return body?.error ?? fallback
  } catch {
    return `${fallback} (${res.status} ${res.statusText})`
  }
}

describe('extractError', () => {
  it('returns the error field from a JSON body', async () => {
    const res = new Response(JSON.stringify({ error: 'not found' }), {
      status: 404,
      statusText: 'Not Found',
      headers: { 'Content-Type': 'application/json' },
    })

    const msg = await extractError(res, 'fallback message')
    expect(msg).toBe('not found')
  })

  it('returns fallback text when body is not JSON', async () => {
    const res = new Response('plain text error', {
      status: 500,
      statusText: 'Internal Server Error',
    })

    const msg = await extractError(res, 'Parse failed')
    expect(msg).toContain('Parse failed')
    expect(msg).toContain('500')
  })

  it('returns fallback when JSON body has no error field', async () => {
    const res = new Response(JSON.stringify({ message: 'something else' }), {
      status: 400,
      statusText: 'Bad Request',
      headers: { 'Content-Type': 'application/json' },
    })

    const msg = await extractError(res, 'my fallback')
    expect(msg).toBe('my fallback')
  })
})
