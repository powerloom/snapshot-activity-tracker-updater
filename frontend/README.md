# DSV Network Activity Dashboard - Frontend

A React-based web dashboard for visualizing decentralized sequencer validator (DSV) network activity.

## Tech Stack

- **React 18** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **TanStack Query (React Query)** - Data fetching and caching
- **D3.js** - Network topology visualization
- **TailwindCSS** - Styling

## Development

### Prerequisites

- Node.js 20+
- npm or yarn

### Setup

```bash
# Install dependencies
npm install

# Start development server (with API proxy)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

The dev server runs on `http://localhost:3000` and proxies API requests to the Go backend at `http://localhost:8080`.

## Project Structure

```
src/
├── api/
│   ├── client.ts       # API client using axios
│   └── types.ts        # TypeScript types matching Go responses
├── components/
│   └── NetworkTopology.tsx  # D3.js force-directed graph
├── pages/
│   └── Dashboard.tsx   # Main dashboard page
├── hooks/
│   └── useDashboardData.ts  # TanStack Query hooks
├── App.tsx
├── main.tsx
└── index.css
```

## API Endpoints

The dashboard connects to the Go backend API:

| Endpoint | Response |
|----------|----------|
| `GET /api/health` | Health check |
| `GET /api/dashboard/summary` | Overview stats |
| `GET /api/network/topology` | Network graph data |
| `GET /api/epochs` | List epochs |
| `GET /api/epochs/:id` | Epoch details |
| `GET /api/validators` | Validator list |
| `GET /api/validators/:id` | Validator details |
| `GET /api/slots` | Slot list |
| `GET /api/slots/:id` | Slot details |
| `GET /api/projects` | Project list |
| `GET /api/timeline` | Timeline of events |

## Data Model

### Network Topology

The network graph displays three types of nodes:

- **Validators** (green) - Validator nodes that aggregate batches
- **Slots** (blue) - Snapshotter slots that submit snapshots
- **Projects** (amber) - Data projects being voted on

Links represent:
- `validates` - Validator validated a project
- `submits_to` - Slot submitted to a project
- `votes_for` - Vote relationships

## Environment Variables

```bash
# API base URL (optional, defaults to /api for proxy)
VITE_API_URL=http://localhost:8080/api
```

## Building

The frontend is built into `dist/` and embedded into the Go binary using `go:embed`:

```bash
npm run build
```

The build output is then copied into the Go container during Docker build.
