/**
 * Nodes API endpoints
 */

import { api } from './client';
import type {
	Node,
	NodeResponse,
	NodesListResponse,
	EdgesListResponse,
	NodeQueryParams,
	NodesListQueryParams
} from './types';

/**
 * Encode section symbols (§) properly in URLs
 */
function encodeNodeId(id: string): string {
	return encodeURIComponent(id);
}

/**
 * Get a single node by ID
 * @param id - Canonical node ID (e.g., "CA:CIV:T02:CH02:§3342")
 * @param params - Optional query params (expand=parents|children)
 */
export async function getNode(id: string, params?: NodeQueryParams): Promise<NodeResponse> {
	return api.get<NodeResponse>(`/nodes/${encodeNodeId(id)}`, params);
}

/**
 * Get child nodes of a node
 * @param id - Canonical node ID
 * @param params - Optional query params (labels, limit, offset, sort, etc.)
 */
export async function getNodeChildren(
	id: string,
	params?: NodesListQueryParams
): Promise<NodesListResponse> {
	return api.get<NodesListResponse>(`/nodes/${encodeNodeId(id)}/children`, params);
}

/**
 * Get parent nodes of a node
 * @param id - Canonical node ID
 * @param params - Optional query params (labels, limit, offset, sort, etc.)
 */
export async function getNodeParents(
	id: string,
	params?: NodesListQueryParams
): Promise<NodesListResponse> {
	return api.get<NodesListResponse>(`/nodes/${encodeNodeId(id)}/parents`, params);
}

/**
 * Get reverse citations (what cites this node)
 * @param id - Canonical node ID
 * @param params - Optional query params (labels, limit, offset, cursor, etc.)
 */
export async function getNodeCitations(
	id: string,
	params?: NodesListQueryParams
): Promise<EdgesListResponse> {
	return api.get<EdgesListResponse>(`/nodes/${encodeNodeId(id)}/citations`, params);
}

/**
 * Get outgoing citations (what this node cites)
 * @param id - Canonical node ID
 * @param params - Optional query params (labels, limit, offset, cursor, etc.)
 */
export async function getNodeCites(
	id: string,
	params?: NodesListQueryParams
): Promise<EdgesListResponse> {
	return api.get<EdgesListResponse>(`/nodes/${encodeNodeId(id)}/cites`, params);
}
