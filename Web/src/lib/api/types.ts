/**
 * Domain types for LawMap API
 * Based on the graph model documented in API/docs/GRAPH_MODEL.md
 */

// Node labels in the graph
export type NodeLabel =
	| 'JURISDICTION'
	| 'CODE'
	| 'TITLE'
	| 'CHAPTER'
	| 'SECTION'
	| 'OPINION'
	| 'RULE'
	| 'REGULATION'
	| 'TOPIC';

// Edge types in the graph
export type EdgeType =
	| 'PARENT_OF'
	| 'AMENDS'
	| 'REPEALS'
	| 'CITES'
	| 'INTERPRETS'
	| 'HAS_TOPIC'
	| 'SAME_AS';

// Version metadata for change tracking
export interface Version {
	fetched_at: string; // ISO 8601 timestamp
	effective_date?: string; // ISO 8601 date
	hash: string; // Content hash for change detection
}

// Source provenance
export interface Source {
	name: string; // e.g., "CA LegInfo", "US Code (OLRC)"
	url: string;
	fetched_at: string;
}

// Graph node
export interface Node {
	id: string; // Canonical ID: jurisdiction:code:title:chapter:section
	labels: NodeLabel[];
	properties: {
		jurisdiction?: string;
		code?: string;
		title?: string;
		chapter?: string;
		section?: string;
		name?: string;
		text?: string;
		summary?: string;
		[key: string]: unknown; // Allow additional properties
	};
	version: Version;
	sources: Source[];
}

// Graph edge
export interface Edge {
	id: string;
	type: EdgeType;
	from: string; // Node ID
	to: string; // Node ID
	properties?: {
		pin_cite?: string;
		context?: string;
		[key: string]: unknown;
	};
}

// Pagination metadata
export interface PaginationMeta {
	total: number;
	limit: number;
	offset: number;
	next_offset?: number;
	next_cursor?: string;
}

// API response wrappers
export interface NodeResponse {
	node: Node;
	parents?: Node[];
	children?: Node[];
}

export interface NodesListResponse {
	nodes: Node[];
	meta: PaginationMeta;
}

export interface EdgesListResponse {
	edges: Edge[];
	nodes: Record<string, Node>; // Map of node IDs to nodes
	meta: PaginationMeta;
}

export interface GraphResponse {
	root: string; // Root node ID
	nodes: Node[];
	edges: Edge[];
	depth: number;
}

export interface SearchResponse {
	results: Node[];
	meta: PaginationMeta;
}

export interface Topic {
	id: string;
	name: string;
	description?: string;
	count: number; // Number of items with this topic
}

export interface TopicsResponse {
	topics: Topic[];
}

export interface SourceInfo {
	name: string;
	jurisdictions: string[];
	codes: string[];
	kind: string;
	enabled: boolean;
	last_fetch?: string;
}

export interface SourcesResponse {
	sources: SourceInfo[];
}

export interface HealthResponse {
	status: 'ok' | 'degraded' | 'down';
	version: string;
	timestamp: string;
}

// Query parameters
export interface NodeQueryParams {
	expand?: 'parents' | 'children' | 'both';
}

export interface NodesListQueryParams {
	labels?: NodeLabel[];
	limit?: number;
	offset?: number;
	cursor?: string;
	sort?: string;
	fields?: string[];
	count_only?: boolean;
}

export interface GraphQueryParams {
	root: string;
	depth?: number;
	labels?: NodeLabel[];
}

export interface SearchQueryParams {
	q: string;
	jurisdiction?: string;
	code?: string;
	labels?: NodeLabel[];
	limit?: number;
	offset?: number;
}

// API Error
export interface ApiError {
	error: string;
	message: string;
	status: number;
}
