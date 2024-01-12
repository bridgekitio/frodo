package services

import (
	"io"
)

// ContentGetter provides a way for your service response to indicate that you want to return a
// raw stream of bytes rather than relying on our auto-encoding.
type ContentGetter interface {
	// Content returns the stream of exact bytes to send back to the caller.
	Content() io.ReadCloser
}

// ContentSetter allows raw stream responses to be properly reconstituted when using the
// code-generated Go client for your service.
type ContentSetter interface {
	// SetContent applies the stream of bytes that the response object should use when reading.
	SetContent(io.ReadCloser)
}

// ContentTypeGetter is used by raw response streams to indicate what type of data is in the stream.
type ContentTypeGetter interface {
	// ContentType returns the MIME encoding type of the raw content stream.
	ContentType() string
}

// ContentTypeSetter recaptures the custom content type header from raw responses when using the
// code-generated Go client for your service.
type ContentTypeSetter interface {
	// SetContentType sets the MIME encoding type of the raw content stream.
	SetContentType(string)
}

// ContentLengthGetter is used by raw response streams to indicate exactly how many bytes are
// in the response's Content stream.
type ContentLengthGetter interface {
	// ContentLength returns the total number of bytes you can/should read from the Content stream.
	ContentLength() int
}

// ContentLengthSetter recaptures the custom content length header from raw responses when using
// the code-generate Go client for your services.
type ContentLengthSetter interface {
	// SetContentLength applies the total number of bytes you can/should read from the Content stream.
	SetContentLength(int)
}

// ContentRangeGetter is used by raw response streams to indicate that this is resumable using the
// standard Range header.
type ContentRangeGetter interface {
	// ContentRange returns the individual values used to build a custom Content-Range header
	// when responding with a raw content stream. For more information on how these values are
	// used, please see: https://www.geeksforgeeks.org/http-headers-content-range/
	ContentRange() (start int, end int, size int)
}

// ContentRangeSetter recaptures the custom content range header from raw responses when using
// the code-generated Go client for your services.
type ContentRangeSetter interface {
	// SetContentRange accepts the 3 standard components of the Content-Range header.
	SetContentRange(start int, end int, size int)
}

// ContentFileNameGetter is used to supply an optional Content-Disposition header, allowing you to customize
// the name of the file presented in download modals of browsers/clients.
type ContentFileNameGetter interface {
	// ContentFileName returns the name of the file that should be used when downloading this stream.
	ContentFileName() string
}

// ContentFileNameSetter is used to supply an optional Content-Disposition header, allowing you to customize
// the name of the file presented in download modals of browsers/clients.
type ContentFileNameSetter interface {
	// SetContentFileName applies the name of the file that should be used when downloading this stream.
	SetContentFileName(string)
}

// StreamRequest implements all of the ContentXxx and SetContentXxx methods that we support and look
// at when we look at streaming/upload style requests.
//
//	type FileUploadRequest struct {
//		services.StreamRequest
//	}
//
//	func (res *ImageDownloadResponse) Init(file os.File, info fs.FileInfo) {
//		res.SetContent(file)
//		res.SetContentType("image/png")
//		res.SetContentLength(info.Size())
//	}
type StreamRequest struct {
	content       io.ReadCloser
	contentType   string
	contentLength int
}

// StreamResponse implements all of the ContentXxx and SetContentXxx methods that we support. You
// can embed one of these structs in your response struct to automatically gain the ability to
// respond with raw data streams rather than auto-encoding.
//
//	type ImageDownloadResponse struct {
//		services.StreamResponse
//	}
//
//	func (res *ImageDownloadResponse) Init(file os.File, info fs.FileInfo) {
//		res.SetContent(file)
//		res.SetContentType("image/png")
//		res.SetContentLength(info.Size())
//	}
type StreamResponse struct {
	content           io.ReadCloser
	contentType       string
	contentLength     int
	contentRangeStart int
	contentRangeEnd   int
	contentRangeSize  int
	contentFileName   string
}

// Content returns the raw byte stream representing the data returned by the endpoint.
func (res *StreamResponse) Content() io.ReadCloser {
	return res.content
}

// SetContent applies the raw byte stream representing the data returned by the endpoint.
func (res *StreamResponse) SetContent(content io.ReadCloser) {
	res.content = content
}

// ContentRange returns non-zero values if this resource supports the ability to resume downloads.
func (res *StreamResponse) ContentRange() (start int, end int, size int) {
	return res.contentRangeStart, res.contentRangeEnd, res.contentRangeSize
}

// SetContentRange applies the attributes related to controlling resumable downloads.
func (res *StreamResponse) SetContentRange(start int, end int, size int) {
	res.contentRangeStart = start
	res.contentRangeEnd = end
	res.contentRangeSize = size
}

// ContentType returns the MIME content type string describe the type of data in the stream.
func (res *StreamResponse) ContentType() string {
	return res.contentType
}

// SetContentType applies the MIME content type that describes the data in the stream.
func (res *StreamResponse) SetContentType(contentType string) {
	res.contentType = contentType
}

// ContentLength returns the number of bytes you can read from the content stream.
func (res *StreamResponse) ContentLength() int {
	return res.contentLength
}

// SetContentLength sets the number of bytes the caller should read from the content stream.
func (res *StreamResponse) SetContentLength(contentLength int) {
	res.contentLength = contentLength
}

// ContentFileName returns the name of the file the client should use to download the stream.
func (res *StreamResponse) ContentFileName() string {
	return res.contentFileName
}

// SetContentFileName sets the name of the file the client should use to download the stream.
func (res *StreamResponse) SetContentFileName(contentFileName string) {
	res.contentFileName = contentFileName
}

// Redirector provides a way to tell gateways that the response value doesn't contain the
// raw byte stream we want to deliver. Instead, you should redirect to that URI to fetch
// the response data.
//
// This indicates the redirect is temporary, and you should probably continue to use the same
// endpoint address in the future. You'd probably use this more in cases such as redirecting
// to a file on S3; something that will be different each time.
//
// GATEWAY COMPATABILITY: This currently only works with the API gateway. When delivering/receiving
// responses through other gateways such as "Events", your response will be auto-encoded just
// like it was a normal struct/value. As a result, your response should continue to maintain
// exported fields that you would like to transport in those cases.
type Redirector interface {
	// Redirect returns the URI of an alternate resource that will provide the final data
	// we want this endpoint to return.
	Redirect() string
}

// RedirectorPermanent provides a way to tell gateways that the response value doesn't contain the
// raw byte stream we want to deliver. Instead, you should redirect to that URI to fetch
// the response data.
//
// This indicates that the redirect is permanent, and you should probably start using the
// redirected URI moving forward. You'd probably use this more in a situation where you are
// deprecating one API endpoint in favor of another. The old endpoint could redirect to the
// new endpoint to maintain backwards compatability, but you really should start using the new one.
//
// GATEWAY COMPATABILITY: This currently only works with the API gateway. When delivering/receiving
// responses through other gateways such as "Events", your response will be auto-encoded just
// like it was a normal struct/value. As a result, your response should continue to maintain
// exported fields that you would like to transport in those cases.
type RedirectorPermanent interface {
	// RedirectPermanent returns the URI of an alternate resource that will provide the final data
	// we want this endpoint to return.
	RedirectPermanent() string
}
