/**
 * Graph traversal and visualization endpoints
 */

import { api } from './client';
import type { GraphResponse, GraphQueryParams } from './types';

/**
 * Get a graph slice with depth control
 * @param params - Query params (root, depth, labels)
 * @example
 * ```ts
 * const graph = await getGraph({
 *   root: 'CA:CIV:T02:CH02:§3342',
 *   depth: 2,
 *   labels: ['SECTION', 'CHAPTER']
 * });
 * ```
 */
export async function getGraph(params: GraphQueryParams): Promise<GraphResponse> {
	return api.get<GraphResponse>('/graph', params);
}

/**
 * Get a subgraph starting from a root node
 * Convenience wrapper for getGraph with better naming
 */
export async function getSubgraph(
	rootId: string,
	depth: number = 2,
	labels?: GraphQueryParams['labels']
): Promise<GraphResponse> {
	return getGraph({ root: rootId, depth, labels });
}
