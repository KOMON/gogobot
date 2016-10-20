package handler

type Handler interface {
	Match(string) bool
	Respond() (string, error)
}

type Router struct {
	Handlers []Handler
}

type handlerError struct {
	s string
}

func (e *handlerError) Error() string {
	return e.s
}

func (r Router) Route(msg string) (*Handler, error) {
	if r.Handlers == nil {
		return nil, &handlerError{"no handlers!"}
	}

	for _, handler := range r.Handlers {
		if handler.Match(msg) {
			return &handler, nil
		}

	}
	return nil, nil
}
