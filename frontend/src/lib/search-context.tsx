import { createContext, useContext, useState, ReactNode } from 'react'

interface SearchContextValue {
  query: string
  setQuery: (q: string) => void
  teamFilter: string
  setTeamFilter: (t: string) => void
  actorFilter: string
  setActorFilter: (a: string) => void
  teamTypeFilter: string
  setTeamTypeFilter: (tt: string) => void
}

const SearchContext = createContext<SearchContextValue | null>(null)

export function SearchProvider({ children }: { children: ReactNode }) {
  const [query, setQuery] = useState('')
  const [teamFilter, setTeamFilter] = useState('')
  const [actorFilter, setActorFilter] = useState('')
  const [teamTypeFilter, setTeamTypeFilter] = useState('')

  return (
    <SearchContext.Provider value={{ query, setQuery, teamFilter, setTeamFilter, actorFilter, setActorFilter, teamTypeFilter, setTeamTypeFilter }}>
      {children}
    </SearchContext.Provider>
  )
}

export function useSearch() {
  const ctx = useContext(SearchContext)
  if (!ctx) throw new Error('useSearch must be used within SearchProvider')
  return ctx
}

/** Filter a label against the current query (case-insensitive substring). */
export function matchesQuery(label: string, query: string): boolean {
  if (!query) return true
  return label.toLowerCase().includes(query.toLowerCase())
}
