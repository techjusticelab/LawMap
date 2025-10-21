# LawMap API Client

TypeScript client for the LawMap Go backend API.

## Structure

```
src/lib/api/
├── index.ts      # Main export file
├── types.ts      # TypeScript types for all API entities
├── client.ts     # Base HTTP client with error handling
├── nodes.ts      # Node CRUD and traversal endpoints
├── graph.ts      # Graph traversal endpoints
└── search.ts     # Search, topics, and sources endpoints
```

## Usage

### Import

```ts
import { getNode, search, getGraph, type Node, type SearchResponse } from '$lib/api';
```

### Get a Node

```ts
// Get a single node
const { node } = await getNode('CA:CIV:T02:CH02:§3342');

// Get node with parents expanded
const { node, parents } = await getNode('CA:CIV:T02:CH02:§3342', { expand: 'parents' });
```

### Get Children/Parents

```ts
// Get children with filtering
const { nodes, meta } = await getNodeChildren('CA:CIV:T02', {
  labels: ['CHAPTER', 'SECTION'],
  limit: 50,
  offset: 0
});

// Get parents
const { nodes, meta } = await getNodeParents('CA:CIV:T02:CH02:§3342');
```

### Citations

```ts
// Get what cites this node (reverse citations)
const { edges, nodes, meta } = await getNodeCitations('CA:CIV:T02:CH02:§3342', {
  limit: 100
});

// Get what this node cites (outgoing citations)
const { edges, nodes, meta } = await getNodeCites('CA:CIV:T02:CH02:§3342');
```

### Graph Traversal

```ts
// Get a graph slice
const graph = await getGraph({
  root: 'CA:CIV:T02:CH02:§3342',
  depth: 2,
  labels: ['SECTION', 'CHAPTER']
});

// Or use the convenience wrapper
const graph = await getSubgraph('CA:CIV:T02:CH02:§3342', 2);
```

### Search

```ts
// Search for nodes
const { results, meta } = await search({
  q: 'dog bite liability',
  jurisdiction: 'CA',
  code: 'CIV',
  limit: 20,
  offset: 0
});
```

### Topics & Metadata

```ts
// Get all topics
const { topics } = await getTopics();

// Get specific topic
const topic = await getTopic('TOPIC:Dogs');

// Get configured sources
const { sources } = await getSources();

// Health check
const health = await getHealth();
```

## Error Handling

All API functions throw `LawMapApiError` on failure:

```ts
import { getNode, LawMapApiError } from '$lib/api';

try {
  const { node } = await getNode('CA:CIV:T02:CH02:§3342');
} catch (error) {
  if (error instanceof LawMapApiError) {
    console.error(`API Error (${error.status}): ${error.message}`);
  } else {
    console.error('Unexpected error:', error);
  }
}
```

## Environment Configuration

Set the API base URL in `.env`:

```env
PUBLIC_API_URL="http://localhost:8080"
```

For production:

```env
PUBLIC_API_URL="https://api.lawmap.example.com"
```

## TypeScript Types

All types are exported from the main module:

```ts
import type {
  Node,
  Edge,
  NodeLabel,
  EdgeType,
  GraphResponse,
  SearchResponse,
  PaginationMeta
} from '$lib/api';
```

### Key Types

- **Node**: Graph node with canonical ID, labels, properties, version, sources
- **Edge**: Graph edge with type, from/to node IDs, optional properties
- **NodeLabel**: Union type of all node labels (`JURISDICTION`, `CODE`, `TITLE`, etc.)
- **EdgeType**: Union type of all edge types (`PARENT_OF`, `CITES`, `AMENDS`, etc.)
- **PaginationMeta**: Pagination metadata with total, limit, offset, cursors

## SvelteKit Integration

### In `+page.server.ts` (Server-side)

```ts
import { getNode } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
  const { node } = await getNode(params.id);
  return { node };
};
```

### In `+page.svelte` (Client-side)

```ts
import { search } from '$lib/api';

let searchQuery = $state('');
let results = $state([]);

async function handleSearch() {
  const response = await search({ q: searchQuery, limit: 20 });
  results = response.results;
}
```

## Canonical ID Format

Node IDs follow the format: `jurisdiction:code:title:chapter:section`

Examples:
- `CA:CIV:T02:CH02:§3342` - California Civil Code, Title 2, Chapter 2, Section 3342
- `US:USC:T18:§924(e)` - US Code, Title 18, Section 924(e)

The `§` symbol is automatically URL-encoded by the client.
