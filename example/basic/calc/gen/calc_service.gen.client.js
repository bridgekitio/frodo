// Code generated by Frodo - DO NOT EDIT.
//
//   Timestamp: Thu, 29 Jun 2023 10:10:18 EDT
//   Source:    calc/calc_service.go
//   Generator: https://github.com/bridgekitio/frodo
//
/* global fetch,module,window */
'use strict';

/**
 * Exposes all of the standard operations for the remote CalculatorService service. These RPC calls
 * will be sent over http(s) to the backend service instances. 
 * CalculatorService provides the ability to perform basic arithmetic on two numbers.
 */
class CalculatorServiceClient {
    _baseURL;
    _fetch;
    _authorization;

    /**
     * @param {string} baseURL The protocol/host/port used by all API/service
     *     calls (e.g. "https://some-server:9000")
     * @param {object} [options]
     * @param {fetch|*} [options.fetch] Provide a custom implementation for the 'fetch' API. Not
     *     necessary if running in browser.
     * @param {string} [options.authorization] Use these credentials in the HTTP Authorization header
     *      for every request. Only use the client-level authorization when all requests to the
     *      service should have the same credentials. If you allow multiple users in your system,
     *      leave this blank and use the authorization option on each request.
     */
    constructor(baseURL, {fetch, authorization} = {}) {
        this._baseURL = trimSlashes(baseURL);
        this._fetch = fetch || defaultFetch();
        this._authorization = authorization || '';
    }
    
    /**
     * Add calculates and returns the sum of two numbers. 
     *
     * @param { AddRequest } serviceRequest The input parameters
     * @param {object} [options]
     * @param { string } [options.authorization] The HTTP Authorization header value to include
     *     in the request. This will override any authorization you might have applied when
     *     constructing this client. Use this in multi-tenant situations where multiple users
     *     might utilize this service.
     * @returns {Promise<AddResponse> } The JSON-encoded return value of the operation.
     */
    async Add(serviceRequest, {authorization} = {}) {
        if (!serviceRequest) {
            throw new GatewayError(400, 'precondition failed: empty request');
        }

        const method = 'GET';
        const route = '/add/{A}/{B}';
        const url = this._baseURL + '/' + buildRequestPath(method, route, serviceRequest);
        const fetchOptions = {
            method: method,
            headers: {
                'Authorization': authorization || this._authorization,
                'Accept': 'application/json,*/*',
                'Content-Type': 'application/json; charset=utf-8',
            },
        };

        const response = await doFetch(this._fetch, url, fetchOptions);
        return handleResponseJSON(response);
    
    }
    
    
    /**
     * Double multiplies the value by 2 
     *
     * @param { DoubleRequest } serviceRequest The input parameters
     * @param {object} [options]
     * @param { string } [options.authorization] The HTTP Authorization header value to include
     *     in the request. This will override any authorization you might have applied when
     *     constructing this client. Use this in multi-tenant situations where multiple users
     *     might utilize this service.
     * @returns {Promise<DoubleResponse> } The JSON-encoded return value of the operation.
     */
    async Double(serviceRequest, {authorization} = {}) {
        if (!serviceRequest) {
            throw new GatewayError(400, 'precondition failed: empty request');
        }

        const method = 'POST';
        const route = '/double/{Value}';
        const url = this._baseURL + '/' + buildRequestPath(method, route, serviceRequest);
        const fetchOptions = {
            method: method,
            headers: {
                'Authorization': authorization || this._authorization,
                'Accept': 'application/json,*/*',
                'Content-Type': 'application/json; charset=utf-8',
            },
            body: JSON.stringify(serviceRequest),
        };

        const response = await doFetch(this._fetch, url, fetchOptions);
        return handleResponseJSON(response);
    
    }
    
    
    /**
     * Mul calculates and returns the product of two numbers. 
     *
     * @param { MulRequest } serviceRequest The input parameters
     * @param {object} [options]
     * @param { string } [options.authorization] The HTTP Authorization header value to include
     *     in the request. This will override any authorization you might have applied when
     *     constructing this client. Use this in multi-tenant situations where multiple users
     *     might utilize this service.
     * @returns {Promise<MulResponse> } The JSON-encoded return value of the operation.
     */
    async Mul(serviceRequest, {authorization} = {}) {
        if (!serviceRequest) {
            throw new GatewayError(400, 'precondition failed: empty request');
        }

        const method = 'GET';
        const route = '/multiply/{A}/{B}';
        const url = this._baseURL + '/' + buildRequestPath(method, route, serviceRequest);
        const fetchOptions = {
            method: method,
            headers: {
                'Authorization': authorization || this._authorization,
                'Accept': 'application/json,*/*',
                'Content-Type': 'application/json; charset=utf-8',
            },
        };

        const response = await doFetch(this._fetch, url, fetchOptions);
        return handleResponseJSON(response);
    
    }
    
    
    /**
     * Sub calculates and returns the difference between two numbers. 
     *
     * @param { SubRequest } serviceRequest The input parameters
     * @param {object} [options]
     * @param { string } [options.authorization] The HTTP Authorization header value to include
     *     in the request. This will override any authorization you might have applied when
     *     constructing this client. Use this in multi-tenant situations where multiple users
     *     might utilize this service.
     * @returns {Promise<SubResponse> } The JSON-encoded return value of the operation.
     */
    async Sub(serviceRequest, {authorization} = {}) {
        if (!serviceRequest) {
            throw new GatewayError(400, 'precondition failed: empty request');
        }

        const method = 'GET';
        const route = '/sub/{A}/{B}';
        const url = this._baseURL + '/' + buildRequestPath(method, route, serviceRequest);
        const fetchOptions = {
            method: method,
            headers: {
                'Authorization': authorization || this._authorization,
                'Accept': 'application/json,*/*',
                'Content-Type': 'application/json; charset=utf-8',
            },
        };

        const response = await doFetch(this._fetch, url, fetchOptions);
        return handleResponseJSON(response);
    
    }
    
}

/**
 * Fills in a router path pattern such as "/user/{id}", with the appropriate attribute from
 * the 'serviceRequest' instance.
 *
 * @param {string} method The HTTP method for this request (determines if we include a query string)
 * @param {string} path The path pattern to populate w/ runtime values (e.g. "/user/{id}")
 * @param {Object} serviceRequest The input struct for the service call
 * @returns {string} The fully-populate URL path (e.g. "/user/aCx31s")
 */
function buildRequestPath(method, path, serviceRequest) {
    const values = new URLValues(serviceRequest);

    const pathSegments = path.split('/').map(segment => {
        return segment.startsWith('{') && segment.endsWith('}')
            ? encodeURIComponent(values.get(segment.substring(1, segment.length - 1)))
            : segment;
    });
    const resolvedPath = trimSlashes(pathSegments.join('/'));

    // PUT/POST/PATCH:  encode the data in the body, so no need to shove it in the query string.
    // GET/DELETE/HEAD: will pass all values through the query string.
    return supportsBody(method)
        ? resolvedPath
        : resolvedPath + '?' + values.format();
}

/**
 * URLValues helps convert a single request object into a map of individual attributes that can
 * be easily added to a path or query string.
 *
 * Example:
 * ```
 * let req = {
 *   ID: '123',
 *   Name: 'Bob',
 *   Alive: true,
 *   ContactInfo: {
 *     PhoneNumber: '123-555-1234',
 *     Emails: { Home: 'me@you.com', Work: 'work@you.com' },
 *   },
 * };
 *
 * let values = URLValues(req);
 * console.info(values.get('ID'));                     // 123
 * console.info(values.get('Alive'));                  // true
 * console.info(values.get('ContactInfo.Email.Home')); // me@you.com
 * ```
 */
class URLValues {
    constructor(struct) {
        this.struct = struct;
        this.attrs = {};
        this._load(struct, '');
    }

    get(name) {
        return this.attrs[name] || '';
    }

    format() {
        const attrs = this.attrs;
        return Object.getOwnPropertyNames(this.attrs)
            .map(attr => attr + '=' + encodeURIComponent(attrs[attr]))
            .join('&');
    }

    _load(value, prefix = '') {
        for (let propertyName of Object.getOwnPropertyNames(value)) {
            let propertyValue = value[propertyName];
            let propertyKey = prefix ? prefix + '.' + propertyName : propertyName;

            if (propertyValue === null) {
                continue;
            }
            switch (typeof propertyValue) {
            case 'undefined':
            case 'symbol':
                continue;

            case 'function':
                this.attrs[propertyKey] = propertyValue();
                continue;

            case 'boolean':
            case 'number':
            case 'bigint':
            case 'string':
                this.attrs[propertyKey] = propertyValue;
                continue;

            default:
                this._load(propertyValue, propertyKey);
            }
        }
    }
}

/**
 * Accepts the full response data and the request's promise resolve/reject and determines
 * which to invoke. This will also JSON-unmarshal the response data if need be.
 */
async function handleResponseJSON(response) {
    if (response.status >= 400) {
        throw await newError(response);
    }
    return await response.json();
}

/**
 * Accepts the full response data and the request's promise resolve/reject and determines
 * which to invoke. This assumes that you want the raw bytes as a blob from the HTTP response
 * rather than treating it like JSON. This will also capture the Content-Type value as well as
 * the "filename" from the Content-Disposition if it's set to "attachment".
 *
 * @returns { StreamedResponse }
 */
async function handleResponseStream(response) {
    if (response.status >= 400) {
        throw await newError(response);
    }
    const content = await response.blob();
    const contentType = response.headers.get('content-type') || 'application/octet-stream';
    const contentFileName = dispositionFileName(response.headers.get('content-disposition'));
    const contentLength = toInt(response.headers.get('content-length')) || 0;
    const contentRange = parseContentRange(response.headers.get('content-range'));

    return {
        Content: content,
        ContentType: contentType,
        ContentLength: contentLength,
        ContentFileName: contentFileName,
        ContentRange: contentRange,
    }
}

/**
 * Accepts the 'Content-Range' header value from a response and parses out all 4 components
 * of the value; the unit, start, end, and size. You'll get back a single object containing
 * all 4 values.
 *
 * @returns {ContentRange}
 */
function parseContentRange(range) {
    range = range && range.trim()
    if (!range) {
        return {Unit: 'bytes', Start: 0, End: 0, Size: 0};
    }

    let matches = range.match(/^(\w*) /);
    const unit = matches && matches[1];

    matches = range.match(/(\d+)-(\d+)\/(\d+|\*)/);
    if (matches) {
        return {
            Unit: unit || 'bytes',
            Start: toInt(matches[1]),
            End: toInt(matches[2]),
            Size: matches[3] === '*' ? null : toInt(matches[3])
        };
    }
}

/**
 * An alternative to the standard 'parseInt' that handles shittier cases like '5x1'. Standard
 * parseInt() returns 5 whereas toInt() returns NaN as you'd expect.
 */
function toInt(value) {
    const num = Number(value);
    return num >= 0 ? Math.floor(num) : Math.ceil(num);
}

/**
 * Creates a new GatewayError with all of the meaningful status/message info extracted
 * from the HTTP response.
 *
 * @returns {Promise< GatewayError >}
 */
async function newError(response) {
    const body = isJSON(response)
        ? await response.json()
        : await response.text();

    // One of the framework's standard status/message errors, already.
    throw (body['Status'] && body['Message'])
        ? new GatewayError(body['Status'], body['Message'])
        : new GatewayError(response.status, parseErrorMessage(body));
}

/**
 * Parses a value from the Content-Disposition header to extract just the filename attribute.
 *
 * @param {string} contentDisposition
 * @returns {string}
 */
function dispositionFileName(contentDisposition = '') {
    contentDisposition = contentDisposition && contentDisposition.trim();
    if (!contentDisposition) {
        return '';
    }

    const fileNameAttrPos = contentDisposition.indexOf('filename=');
    if (fileNameAttrPos < 0) {
        return '';
    }

    let fileName = contentDisposition.substring(fileNameAttrPos + 9);
    fileName = fileName.startsWith('"') ? fileName.substring(1) : fileName;
    fileName = fileName.endsWith('"') ? fileName.substring(0, fileName.length - 1) : fileName;
    fileName = fileName.replace(/\\"/g, '"');
    return fileName;
}

/**
 * Determines whether or not the response has a content type of JSON.
 */
function isJSON(response) {
    const contentType = response.headers.get('content-type');
    return contentType && contentType.toLowerCase().startsWith('application/json');
}

/**
* Looks at the response value and attempts to peel off an error message from it using the standard
* error JSON structures used by frodo gateways.
*
* @param {*} err The error whose raw message you're trying to extract.
* @returns {string}
*/
function parseErrorMessage(err) {
    if (typeof err === 'string') {
        return err;
    }
    if (typeof err.message !== 'undefined') {
        return err.message;
    }
    if (typeof err.error !== 'undefined') {
        return err.error;
    }
    return JSON.stringify(err);
}

/**
 * Does the HTTP method given support supplying data in the body of the request? For instance
 * this is true for POST but not for GET.
 *
 * @param {string} method The HTTP method that you are processing (e.g. "GET", "POST", etc.)
 * @returns {boolean}
 */
function supportsBody(method) {
    return method === 'POST' || method === 'PUT' || method === 'PATCH';
}

/**
 * Removes all leading/trailing slashes from the given URL segment.
 *
 * @param {string} value The URL path segment to clean up.
 * @returns {string}
 */
function trimSlashes(value) {
    if (!value) {
        return "";
    }
    while (value.startsWith("/")) {
        value = value.substring(1);
    }
    while (value.endsWith("/")) {
        value = value.substring(0, value.length - 1);
    }
    return value;
}

/**
* When you don't supply your own Fetch implementation, this will attempt to use
* any globally defined ones (typically for use in the browser).
*
* @returns {fetch}
*/
function defaultFetch() {
    const runningInBrowser = typeof window !== 'undefined';

    if (typeof fetch === 'undefined') {
        throw runningInBrowser
            ? new GatewayError(400, 'no global "fetch" found - unsupported browser')
            : new GatewayError(400, 'no global "fetch" found - upgrade to Node 18+ or install/import node-fetch');
    }
    return runningInBrowser ? fetch.bind(window) : fetch;
}

/**
 * Dispatches your 'fetch' request to the server. Any low-level connection failures will be
 * wrapped in a GatewayError, so we have consistent status codes to better handle error categories.
 */
async function doFetch(fetchFunc, url, options) {
    try {
        return await fetchFunc(url, options);
    } catch (e) {
        throw new GatewayError(502, e.toString());
    }
}


/**
* GatewayError is a rich error type that encapsulates a failure generated by the remote gateway.
* It captures the server's error message as well as HTTP status so you can properly handle the
* result in your consumer code.
*/
class GatewayError {
    /**
    * The HTTP 4XX/5XX status code of the failure.
    *
    * @type {number}
    */
    status;

    /**
    * The user-facing message that the server generated for the error.
    *
    * @type {string}
    */
    message;

    constructor(status, message) {
        this.Status = this.status = status || 500;
        this.Message = this.message = message;
    }

    toString() {
        return this.status + ": " + this.message;
    }
}



/**
 * @typedef { object } MulResponse
 * @property { number|* } [Value]
*/
/**
 * @typedef { object } SubRequest
 * @property { number|* } [A]
 * @property { number|* } [B]
*/
/**
 * @typedef { object } SubResponse
 * @property { number|* } [Value]
*/
/**
 * @typedef { object } DoubleRequest
 * @property { number|* } [Value]
*/
/**
 * @typedef { object } DoubleResponse
 * @property { number|* } [Value]
*/
/**
 * @typedef { object } MulRequest
 * @property { number|* } [A]
 * @property { number|* } [B]
*/
/**
 * @typedef { object } AddResponse
 * @property { number|* } [Value]
*/
/**
 * @typedef { object } AddRequest
 * @property { number|* } [A]
 * @property { number|* } [B]
*/

/**
 * @typedef StreamedResponse
 * @property { Blob } Content
 * @property { string } ContentType
 * @property { number } ContentLength
 * @property { ContentRange } ContentRange
*/

/**
 * @typedef ContentRange
 * @property { string } [Unit]
 * @property { number } [Start]
 * @property { number } [End]
 * @property { number } [Size]
 */

module.exports = {
    CalculatorServiceClient,
};
