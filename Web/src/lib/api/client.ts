/**
 * Base API client with error handling
 */

import { PUBLIC_API_URL } from '$env/static/public';
import type { ApiError } from './types';

export class LawMapApiError extends Error {
	constructor(
		public status: number,
		public error: string,
		message: string
	) {
		super(message);
		this.name = 'LawMapApiError';
	}
}

/**
 * Build URL with query parameters
 */
function buildUrl(path: string, params?: Record<string, unknown>): string {
	const url = new URL(path, PUBLIC_API_URL);

	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined && value !== null) {
				if (Array.isArray(value)) {
					// Handle array parameters (e.g., labels=CODE&labels=SECTION)
					value.forEach((v) => url.searchParams.append(key, String(v)));
				} else {
					url.searchParams.set(key, String(value));
				}
			}
		});
	}

	return url.toString();
}

/**
 * Base fetch wrapper with error handling
 */
async function apiFetch<T>(
	path: string,
	options?: RequestInit & { params?: Record<string, unknown> }
): Promise<T> {
	const { params, ...fetchOptions } = options || {};
	const url = buildUrl(path, params);

	try {
		const response = await fetch(url, {
			...fetchOptions,
			headers: {
				'Content-Type': 'application/json',
				...fetchOptions.headers
			}
		});

		if (!response.ok) {
			let errorData: ApiError;
			try {
				errorData = await response.json();
			} catch {
				errorData = {
					status: response.status,
					error: response.statusText,
					message: `HTTP ${response.status}: ${response.statusText}`
				};
			}

			throw new LawMapApiError(errorData.status, errorData.error, errorData.message);
		}

		return await response.json();
	} catch (error) {
		if (error instanceof LawMapApiError) {
			throw error;
		}

		// Network or parsing error
		throw new LawMapApiError(
			0,
			'NetworkError',
			error instanceof Error ? error.message : 'Unknown error occurred'
		);
	}
}

/**
 * HTTP methods
 */
export const api = {
	get: <T>(path: string, params?: Record<string, unknown>): Promise<T> =>
		apiFetch<T>(path, { method: 'GET', params }),

	post: <T>(path: string, body?: unknown, params?: Record<string, unknown>): Promise<T> =>
		apiFetch<T>(path, {
			method: 'POST',
			body: body ? JSON.stringify(body) : undefined,
			params
		}),

	put: <T>(path: string, body?: unknown, params?: Record<string, unknown>): Promise<T> =>
		apiFetch<T>(path, {
			method: 'PUT',
			body: body ? JSON.stringify(body) : undefined,
			params
		}),

	delete: <T>(path: string, params?: Record<string, unknown>): Promise<T> =>
		apiFetch<T>(path, { method: 'DELETE', params })
};
