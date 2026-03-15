type HeadersWithSetCookie = Headers & {
	getSetCookie?: () => string[];
};

export function authProxyHeaders(request: Request, initialHeaders?: HeadersInit): Headers {
	const headers = new Headers(initialHeaders);
	const cookieHeader = request.headers.get('cookie');
	if (cookieHeader) {
		headers.set('cookie', cookieHeader);
	}

	return headers;
}

function proxyResponseHeaders(sourceHeaders: Headers): Headers {
	const headers = new Headers(sourceHeaders);
	const headersWithSetCookie = sourceHeaders as HeadersWithSetCookie;
	if (typeof headersWithSetCookie.getSetCookie === 'function') {
		const setCookies = headersWithSetCookie.getSetCookie();
		if (setCookies.length > 0) {
			headers.delete('set-cookie');
			for (const setCookie of setCookies) {
				headers.append('set-cookie', setCookie);
			}
		}
	}

	return headers;
}

export function proxiedResponse(response: Response): Response {
	return new Response(response.body, {
		status: response.status,
		statusText: response.statusText,
		headers: proxyResponseHeaders(response.headers)
	});
}
