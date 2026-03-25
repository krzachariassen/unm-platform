import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { type ReactNode } from 'react'
import { ModelProvider, useModel } from './model-context'
import type { ParseResponse } from './api'

const wrapper = ({ children }: { children: ReactNode }) => (
  <ModelProvider>{children}</ModelProvider>
)

const mockParseResponse: ParseResponse = {
  id: 'test-model-123',
  system_name: 'Test System',
  system_description: 'A test system',
  summary: {
    actors: 1,
    needs: 1,
    capabilities: 1,
    services: 1,
    teams: 1,
  },
  validation: {
    is_valid: true,
    errors: [],
    warnings: [],
  },
}

// Mock localStorage for tests
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value }),
    removeItem: vi.fn((key: string) => { delete store[key] }),
    clear: vi.fn(() => { store = {} }),
  }
})()

vi.stubGlobal('localStorage', localStorageMock)

beforeEach(() => {
  localStorageMock.clear()
  vi.clearAllMocks()
  // Re-stub after clearing mocks
  vi.stubGlobal('localStorage', localStorageMock)
})

describe('ModelProvider', () => {
  it('setModel persists model id to localStorage', () => {
    const { result } = renderHook(() => useModel(), { wrapper })

    act(() => {
      result.current.setModel('test-model-123', mockParseResponse)
    })

    expect(localStorageMock.setItem).toHaveBeenCalledWith('unm_model_id', 'test-model-123')
  })

  it('clearModel removes model id from localStorage', () => {
    const { result } = renderHook(() => useModel(), { wrapper })

    act(() => {
      result.current.setModel('test-model-123', mockParseResponse)
    })
    act(() => {
      result.current.clearModel()
    })

    expect(localStorageMock.removeItem).toHaveBeenCalledWith('unm_model_id')
    expect(localStorageMock.removeItem).toHaveBeenCalledWith('unm_parse_result')
    expect(localStorageMock.removeItem).toHaveBeenCalledWith('unm_loaded_at')
  })

  it('restores model from localStorage on mount', async () => {
    ;(localStorageMock.getItem as ReturnType<typeof vi.fn>).mockImplementation((key: string): string | null => {
      if (key === 'unm_model_id') return 'stored-id'
      if (key === 'unm_parse_result') return JSON.stringify(mockParseResponse)
      if (key === 'unm_loaded_at') return new Date().toISOString()
      return null
    })

    const { result } = renderHook(() => useModel(), { wrapper })

    // Wait for useEffect to run
    await act(async () => {})

    expect(result.current.modelId).toBe('stored-id')
  })
})
