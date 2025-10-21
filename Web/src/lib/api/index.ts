/**
 * LawMap API Client
 *
 * Provides typed HTTP client functions for the LawMap Go backend API.
 *
 * @example Basic usage
 * ```ts
 * import { getNode, search } from '$lib/api';
 *
 * // Get a specific node
 * const node = await getNode('CA:CIV:T02:CH02:§3342');
 *
 * // Search
 * const results = await search({ q: 'dog bite', jurisdiction: 'CA' });
 * ```
 */

// Export all types
export type * from './types';

// Export error class
export { LawMapApiError } from './client';

// Export node operations
export {
	getNode,
	getNodeChildren,
	getNodeParents,
	getNodeCitations,
	getNodeCites
} from './nodes';

// Export graph operations
export { getGraph, getSubgraph } from './graph';

// Export search and metadata operations
export { search, getTopics, getTopic, getSources, getHealth } from './search';
