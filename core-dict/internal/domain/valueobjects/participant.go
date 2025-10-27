package valueobjects

import (
	"errors"
	"regexp"
)

// Participant representa um participante do DICT (instituição financeira)
type Participant struct {
	ISPB string // Código ISPB (8 dígitos)
	Name string // Nome da instituição
}

var ispbRegex = regexp.MustCompile(`^\d{8}$`)

// NewParticipant cria e valida um participante
func NewParticipant(ispb, name string) (Participant, error) {
	if !ispbRegex.MatchString(ispb) {
		return Participant{}, errors.New("invalid ISPB: must be 8 digits")
	}
	if name == "" {
		return Participant{}, errors.New("participant name cannot be empty")
	}
	return Participant{
		ISPB: ispb,
		Name: name,
	}, nil
}

// Equals verifica se dois participantes são iguais
func (p Participant) Equals(other Participant) bool {
	return p.ISPB == other.ISPB
}

// String retorna a representação string do participante
func (p Participant) String() string {
	return p.ISPB + " - " + p.Name
}
