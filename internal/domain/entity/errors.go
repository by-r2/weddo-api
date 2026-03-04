package entity

import "errors"

var (
	ErrNotFound      = errors.New("recurso não encontrado")
	ErrAlreadyExists = errors.New("recurso já existe")
	ErrUnauthorized  = errors.New("credenciais inválidas")
	ErrInactive      = errors.New("recurso inativo")
)
