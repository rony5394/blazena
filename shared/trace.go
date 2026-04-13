package shared

import (
	"log/slog"

	"github.com/google/uuid"
);

func NewTraceId()string{
	return uuid.New().String();
}

func helper(name string, id *string)slog.Attr{
	if id == nil{
		return slog.String(name, NewTraceId());
	}

	return  slog.String(name, *id);
}

func NewSlogTrace(id *string)slog.Attr{
	return helper("trace", id);
}

func NewSlogOperation(id *string)slog.Attr{
	return helper("operation", id);
}

