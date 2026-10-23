package status

import (
	"option"
	"strconv"
)

type Code union {
	Continue
	SwitchingProtocols
	Processing
	EarlyHints
	OK
	Created
	Accepted
	NonAuthoritativeInfo
	NoContent
	ResetContent
	PartialContent
	MultiStatus
	AlreadyReported
	IMUsed
	MultipleChoices
	MovedPermanently
	Found
	SeeOther
	NotModified
	UseProxy
	TemporaryRedirect
	PermanentRedirect
	BadRequest
	Unauthorized
	PaymentRequired
	Forbidden
	NotFound
	MethodNotAllowed
	NotAcceptable
	ProxyAuthRequired
	RequestTimeout
	Conflict
	Gone
	LengthRequired
	PreconditionFailed
	RequestEntityTooLarge
	RequestURITooLong
	UnsupportedMediaType
	RequestedRangeNotSatisfiable
	ExpectationFailed
	Teapot
	MisdirectedRequest
	UnprocessableEntity
	Locked
	FailedDependency
	TooEarly
	UpgradeRequired
	PreconditionRequired
	TooManyRequests
	RequestHeaderFieldsTooLarge
	UnavailableForLegalReasons
	InternalServerError
	NotImplemented
	BadGateway
	ServiceUnavailable
	GatewayTimeout
	HTTPVersionNotSupported
	VariantAlsoNegotiates
	InsufficientStorage
	LoopDetected
	NotExtended
	NetworkAuthenticationRequired
	Other                         int
}

func Parse(code int) Code {
	switch code {
	case 100:
		return Code.Continue()
	case 101:
		return Code.SwitchingProtocols()
	case 102:
		return Code.Processing()
	case 103:
		return Code.EarlyHints()
	case 200:
		return Code.OK()
	case 201:
		return Code.Created()
	case 202:
		return Code.Accepted()
	case 203:
		return Code.NonAuthoritativeInfo()
	case 204:
		return Code.NoContent()
	case 205:
		return Code.ResetContent()
	case 206:
		return Code.PartialContent()
	case 207:
		return Code.MultiStatus()
	case 208:
		return Code.AlreadyReported()
	case 226:
		return Code.IMUsed()
	case 300:
		return Code.MultipleChoices()
	case 301:
		return Code.MovedPermanently()
	case 302:
		return Code.Found()
	case 303:
		return Code.SeeOther()
	case 304:
		return Code.NotModified()
	case 305:
		return Code.UseProxy()
	case 307:
		return Code.TemporaryRedirect()
	case 308:
		return Code.PermanentRedirect()
	case 400:
		return Code.BadRequest()
	case 401:
		return Code.Unauthorized()
	case 402:
		return Code.PaymentRequired()
	case 403:
		return Code.Forbidden()
	case 404:
		return Code.NotFound()
	case 405:
		return Code.MethodNotAllowed()
	case 406:
		return Code.NotAcceptable()
	case 407:
		return Code.ProxyAuthRequired()
	case 408:
		return Code.RequestTimeout()
	case 409:
		return Code.Conflict()
	case 410:
		return Code.Gone()
	case 411:
		return Code.LengthRequired()
	case 412:
		return Code.PreconditionFailed()
	case 413:
		return Code.RequestEntityTooLarge()
	case 414:
		return Code.RequestURITooLong()
	case 415:
		return Code.UnsupportedMediaType()
	case 416:
		return Code.RequestedRangeNotSatisfiable()
	case 417:
		return Code.ExpectationFailed()
	case 418:
		return Code.Teapot()
	case 421:
		return Code.MisdirectedRequest()
	case 422:
		return Code.UnprocessableEntity()
	case 423:
		return Code.Locked()
	case 424:
		return Code.FailedDependency()
	case 425:
		return Code.TooEarly()
	case 426:
		return Code.UpgradeRequired()
	case 428:
		return Code.PreconditionRequired()
	case 429:
		return Code.TooManyRequests()
	case 431:
		return Code.RequestHeaderFieldsTooLarge()
	case 451:
		return Code.UnavailableForLegalReasons()
	case 500:
		return Code.InternalServerError()
	case 501:
		return Code.NotImplemented()
	case 502:
		return Code.BadGateway()
	case 503:
		return Code.ServiceUnavailable()
	case 504:
		return Code.GatewayTimeout()
	case 505:
		return Code.HTTPVersionNotSupported()
	case 506:
		return Code.VariantAlsoNegotiates()
	case 507:
		return Code.InsufficientStorage()
	case 508:
		return Code.LoopDetected()
	case 510:
		return Code.NotExtended()
	case 511:
		return Code.NetworkAuthenticationRequired()
	default:
		return Code.Other(code)
	}
}

func (c Code) Int() int {
	switch value := c.(variant) {
	case Continue:
		return 100
	case SwitchingProtocols:
		return 101
	case Processing:
		return 102
	case EarlyHints:
		return 103
	case OK:
		return 200
	case Created:
		return 201
	case Accepted:
		return 202
	case NonAuthoritativeInfo:
		return 203
	case NoContent:
		return 204
	case ResetContent:
		return 205
	case PartialContent:
		return 206
	case MultiStatus:
		return 207
	case AlreadyReported:
		return 208
	case IMUsed:
		return 226
	case MultipleChoices:
		return 300
	case MovedPermanently:
		return 301
	case Found:
		return 302
	case SeeOther:
		return 303
	case NotModified:
		return 304
	case UseProxy:
		return 305
	case TemporaryRedirect:
		return 307
	case PermanentRedirect:
		return 308
	case BadRequest:
		return 400
	case Unauthorized:
		return 401
	case PaymentRequired:
		return 402
	case Forbidden:
		return 403
	case NotFound:
		return 404
	case MethodNotAllowed:
		return 405
	case NotAcceptable:
		return 406
	case ProxyAuthRequired:
		return 407
	case RequestTimeout:
		return 408
	case Conflict:
		return 409
	case Gone:
		return 410
	case LengthRequired:
		return 411
	case PreconditionFailed:
		return 412
	case RequestEntityTooLarge:
		return 413
	case RequestURITooLong:
		return 414
	case UnsupportedMediaType:
		return 415
	case RequestedRangeNotSatisfiable:
		return 416
	case ExpectationFailed:
		return 417
	case Teapot:
		return 418
	case MisdirectedRequest:
		return 421
	case UnprocessableEntity:
		return 422
	case Locked:
		return 423
	case FailedDependency:
		return 424
	case TooEarly:
		return 425
	case UpgradeRequired:
		return 426
	case PreconditionRequired:
		return 428
	case TooManyRequests:
		return 429
	case RequestHeaderFieldsTooLarge:
		return 431
	case UnavailableForLegalReasons:
		return 451
	case InternalServerError:
		return 500
	case NotImplemented:
		return 501
	case BadGateway:
		return 502
	case ServiceUnavailable:
		return 503
	case GatewayTimeout:
		return 504
	case HTTPVersionNotSupported:
		return 505
	case VariantAlsoNegotiates:
		return 506
	case InsufficientStorage:
		return 507
	case LoopDetected:
		return 508
	case NotExtended:
		return 510
	case NetworkAuthenticationRequired:
		return 511
	case Other:
		return value
	}
}

func (c Code) Text() option.Option[string] {
	switch c.Int() {
	case 100:
		return option.Option.Some("Continue")
	case 101:
		return option.Option.Some("Switching Protocols")
	case 102:
		return option.Option.Some("Processing")
	case 103:
		return option.Option.Some("Early Hints")
	case 200:
		return option.Option.Some("OK")
	case 201:
		return option.Option.Some("Created")
	case 202:
		return option.Option.Some("Accepted")
	case 203:
		return option.Option.Some("Non-Authoritative Information")
	case 204:
		return option.Option.Some("No Content")
	case 205:
		return option.Option.Some("Reset Content")
	case 206:
		return option.Option.Some("Partial Content")
	case 207:
		return option.Option.Some("Multi-Status")
	case 208:
		return option.Option.Some("Already Reported")
	case 226:
		return option.Option.Some("IM Used")
	case 300:
		return option.Option.Some("Multiple Choices")
	case 301:
		return option.Option.Some("Moved Permanently")
	case 302:
		return option.Option.Some("Found")
	case 303:
		return option.Option.Some("See Other")
	case 304:
		return option.Option.Some("Not Modified")
	case 305:
		return option.Option.Some("Use Proxy")
	case 307:
		return option.Option.Some("Temporary Redirect")
	case 308:
		return option.Option.Some("Permanent Redirect")
	case 400:
		return option.Option.Some("Bad Request")
	case 401:
		return option.Option.Some("Unauthorized")
	case 402:
		return option.Option.Some("Payment Required")
	case 403:
		return option.Option.Some("Forbidden")
	case 404:
		return option.Option.Some("Not Found")
	case 405:
		return option.Option.Some("Method Not Allowed")
	case 406:
		return option.Option.Some("Not Acceptable")
	case 407:
		return option.Option.Some("Proxy Authentication Required")
	case 408:
		return option.Option.Some("Request Timeout")
	case 409:
		return option.Option.Some("Conflict")
	case 410:
		return option.Option.Some("Gone")
	case 411:
		return option.Option.Some("Length Required")
	case 412:
		return option.Option.Some("Precondition Failed")
	case 413:
		return option.Option.Some("Request Entity Too Large")
	case 414:
		return option.Option.Some("Request URI Too Long")
	case 415:
		return option.Option.Some("Unsupported Media Type")
	case 416:
		return option.Option.Some("Requested Range Not Satisfiable")
	case 417:
		return option.Option.Some("Expectation Failed")
	case 418:
		return option.Option.Some("I'm a teapot")
	case 421:
		return option.Option.Some("Misdirected Request")
	case 422:
		return option.Option.Some("Unprocessable Entity")
	case 423:
		return option.Option.Some("Locked")
	case 424:
		return option.Option.Some("Failed Dependency")
	case 425:
		return option.Option.Some("Too Early")
	case 426:
		return option.Option.Some("Upgrade Required")
	case 428:
		return option.Option.Some("Precondition Required")
	case 429:
		return option.Option.Some("Too Many Requests")
	case 431:
		return option.Option.Some("Request Header Fields Too Large")
	case 451:
		return option.Option.Some("Unavailable For Legal Reasons")
	case 500:
		return option.Option.Some("Internal Server Error")
	case 501:
		return option.Option.Some("Not Implemented")
	case 502:
		return option.Option.Some("Bad Gateway")
	case 503:
		return option.Option.Some("Service Unavailable")
	case 504:
		return option.Option.Some("Gateway Timeout")
	case 505:
		return option.Option.Some("HTTP Version Not Supported")
	case 506:
		return option.Option.Some("Variant Also Negotiates")
	case 507:
		return option.Option.Some("Insufficient Storage")
	case 508:
		return option.Option.Some("Loop Detected")
	case 510:
		return option.Option.Some("Not Extended")
	case 511:
		return option.Option.Some("Network Authentication Required")
	default:
		return option.Option[string].None()
	}
}

func (c Code) String() string {
	code := strconv.Itoa(c.Int())
	return option.Map(c.Text(), func(text string) string { return code + " " + text }).UnwrapOr(code)
}
