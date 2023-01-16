package pam

// TracingRequest interface simplifies working with Tracing from *Request objects using generics.
type TracingRequest interface {
	RefreshSessionRequest |
		GetSessionRequest |
		GetTransactionsRequest |
		AddTransactionRequest |
		GetGameRoundRequest |
		GetBalanceRequest

	Traceparent() *Traceparent
	Tracestate() *Tracestate
}

func (r RefreshSessionRequest) Traceparent() *Traceparent {
	return r.Params.Traceparent
}

func (r RefreshSessionRequest) Tracestate() *Tracestate {
	return r.Params.Tracestate
}

func (r GetSessionRequest) Traceparent() *Traceparent {
	return r.Params.Traceparent
}

func (r GetSessionRequest) Tracestate() *Tracestate {
	return r.Params.Tracestate
}

func (r GetTransactionsRequest) Traceparent() *Traceparent {
	return r.Params.Traceparent
}

func (r GetTransactionsRequest) Tracestate() *Tracestate {
	return r.Params.Tracestate
}

func (r AddTransactionRequest) Traceparent() *Traceparent {
	return r.Params.Traceparent
}

func (r AddTransactionRequest) Tracestate() *Tracestate {
	return r.Params.Tracestate
}

func (r GetGameRoundRequest) Traceparent() *Traceparent {
	return r.Params.Traceparent
}

func (r GetGameRoundRequest) Tracestate() *Tracestate {
	return r.Params.Tracestate
}

func (r GetBalanceRequest) Traceparent() *Traceparent {
	return r.Params.Traceparent
}

func (r GetBalanceRequest) Tracestate() *Tracestate {
	return r.Params.Tracestate
}
