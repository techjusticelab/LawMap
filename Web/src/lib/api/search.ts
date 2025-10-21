/**
 * Search, topics, and sources endpoints
 */

import { api } from './client';
import type {
	SearchResponse,
	SearchQueryParams,
	TopicsResponse,
	Topic,
	SourcesResponse,
	HealthResponse
} from './types';

/**
 * Search for nodes by text query
 * @param params - Search query params (q, jurisdiction, code, labels, limit, offset)
 * @example
 * ```ts
 * const results = await search({
 *   q: 'dog bite liability',
 *   jurisdiction: 'CA',
 *   code: 'CIV',
 *   limit: 20
 * });
 * ```
 */
export async function search(params: SearchQueryParams): Promise<SearchResponse> {
	return api.get<SearchResponse>('/search', params);
}

/**
 * Get all topics
 */
export async function getTopics(): Promise<TopicsResponse> {
	return api.get<TopicsResponse>('/topics');
}

/**
 * Get a specific topic by ID
 * @param id - Topic ID (e.g., "TOPIC:Dogs")
 */
export async function getTopic(id: string): Promise<Topic> {
	return api.get<Topic>(`/topics/${encodeURIComponent(id)}`);
}

/**
 * Get all configured sources
 */
export async function getSources(): Promise<SourcesResponse> {
	return api.get<SourcesResponse>('/sources');
}

/**
 * Health check endpoint
 */
export async function getHealth(): Promise<HealthResponse> {
	return api.get<HealthResponse>('/health');
}
