# Technology Stack

| Layer | Technology |
|-------|-----------|
| Backend language | Go (latest stable) |
| Backend HTTP | net/http stdlib mux (Go 1.22+ method routing) |
| Backend testing | Go testing + testify (assertions) |
| YAML parsing | gopkg.in/yaml.v3 |
| Config | koanf v2 |
| Frontend framework | React 19 + TypeScript |
| Frontend build | Vite 6 |
| Frontend styling | Tailwind CSS v4 |
| Frontend UI | shadcn/ui (Radix primitives, copy-paste pattern) |
| Frontend icons | Lucide React |
| Frontend routing | React Router |
| Frontend data fetching | TanStack Query (React Query) |
| Frontend graph viz | React Flow (`@xyflow/react`) — for UNM Map and any diagram views |
| API | REST (JSON), frontend proxied to Go backend |

## Notes

- **TanStack Query** replaces manual `useEffect` + `useState` patterns for all server state
- **React Flow** is installed (`@xyflow/react` in `package.json`) and is the required tool for graph visualization — do NOT hand-roll SVG layout engines
- **D3.js** is NOT installed and NOT used — do not reference it
- **shadcn/ui** components are copy-pasted into `components/ui/` (not npm-installed as a package)
