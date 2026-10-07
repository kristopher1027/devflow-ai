package anthropic

import "context"

type unavailableGenerator struct {
	err error
}

// NewUnavailableGenerator returns a TextGenerator that always fails with
// err. It lets the server start without an API key and report "AI is not
// configured" only when someone actually asks for an explanation.
func NewUnavailableGenerator(err error) TextGenerator {
	return &unavailableGenerator{err: err}
}

func (g *unavailableGenerator) Generate(
	ctx context.Context,
	system string,
	prompt string,
) (string, error) {
	return "", g.err
}
