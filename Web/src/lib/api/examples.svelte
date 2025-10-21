<!--
  Example usage of the LawMap API client in Svelte 5

  This file demonstrates common patterns for using the API client.
  Copy these examples into your actual components as needed.
-->

<script lang="ts">
	import { getNode, search, getSubgraph, LawMapApiError } from '$lib/api';
	import type { Node, SearchResponse } from '$lib/api';

	// Example 1: Load a single node
	let nodeId = $state('CA:CIV:T02:CH02:§3342');
	let node = $state<Node | null>(null);
	let nodeError = $state<string | null>(null);

	async function loadNode() {
		try {
			nodeError = null;
			const response = await getNode(nodeId, { expand: 'parents' });
			node = response.node;
		} catch (error) {
			if (error instanceof LawMapApiError) {
				nodeError = `Error ${error.status}: ${error.message}`;
			} else {
				nodeError = 'Unknown error occurred';
			}
		}
	}

	// Example 2: Search with reactive state
	let searchQuery = $state('');
	let searchResults = $state<SearchResponse | null>(null);
	let isSearching = $state(false);

	async function performSearch() {
		if (!searchQuery.trim()) return;

		isSearching = true;
		try {
			searchResults = await search({
				q: searchQuery,
				jurisdiction: 'CA',
				limit: 20
			});
		} catch (error) {
			console.error('Search failed:', error);
		} finally {
			isSearching = false;
		}
	}

	// Example 3: Load graph data for visualization
	let graphData = $state<Awaited<ReturnType<typeof getSubgraph>> | null>(null);

	async function loadGraph(rootId: string, depth = 2) {
		try {
			graphData = await getSubgraph(rootId, depth, ['SECTION', 'CHAPTER']);
		} catch (error) {
			console.error('Failed to load graph:', error);
		}
	}

	// Example 4: Pagination handling
	let currentPage = $state(0);
	let pageSize = $state(20);

	async function loadPage(page: number) {
		const offset = page * pageSize;
		const results = await search({
			q: searchQuery,
			limit: pageSize,
			offset
		});

		if (results.meta.total > 0) {
			currentPage = page;
			searchResults = results;
		}
	}
</script>

<!-- Example 1: Node display -->
<div class="node-loader">
	<input type="text" bind:value={nodeId} placeholder="Enter node ID" />
	<button onclick={loadNode}>Load Node</button>

	{#if nodeError}
		<p class="error">{nodeError}</p>
	{:else if node}
		<div class="node-display">
			<h2>{node.properties.name || node.id}</h2>
			<p>{node.properties.text}</p>
			<small>Labels: {node.labels.join(', ')}</small>
		</div>
	{/if}
</div>

<!-- Example 2: Search interface -->
<div class="search-interface">
	<input
		type="search"
		bind:value={searchQuery}
		placeholder="Search California law..."
		onkeydown={(e) => e.key === 'Enter' && performSearch()}
	/>
	<button onclick={performSearch} disabled={isSearching}>
		{isSearching ? 'Searching...' : 'Search'}
	</button>

	{#if searchResults}
		<div class="results">
			<p>{searchResults.meta.total} results found</p>
			{#each searchResults.results as result}
				<div class="result-item">
					<h3>{result.properties.name || result.id}</h3>
					<p>{result.properties.summary}</p>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Example 3: Graph visualization placeholder -->
<div class="graph-viz">
	<button onclick={() => loadGraph('CA:CIV:T02', 3)}>Load Graph</button>

	{#if graphData}
		<div class="graph-stats">
			<p>Nodes: {graphData.nodes.length}</p>
			<p>Edges: {graphData.edges.length}</p>
			<p>Depth: {graphData.depth}</p>
		</div>
		<!-- Add your visualization library here (D3, Cytoscape, etc.) -->
	{/if}
</div>

<!-- Example 4: Pagination -->
{#if searchResults && searchResults.meta.total > pageSize}
	<div class="pagination">
		<button onclick={() => loadPage(currentPage - 1)} disabled={currentPage === 0}>
			Previous
		</button>
		<span>
			Page {currentPage + 1} of {Math.ceil(searchResults.meta.total / pageSize)}
		</span>
		<button
			onclick={() => loadPage(currentPage + 1)}
			disabled={(currentPage + 1) * pageSize >= searchResults.meta.total}
		>
			Next
		</button>
	</div>
{/if}

<style>
	.error {
		color: red;
	}

	.node-display,
	.result-item {
		border: 1px solid #ccc;
		padding: 1rem;
		margin: 0.5rem 0;
		border-radius: 4px;
	}

	input,
	button {
		padding: 0.5rem;
		margin: 0.25rem;
	}

	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
