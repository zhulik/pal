package pal

var (
	defaultRunConfig = &RunConfig{
		Wait: true,
	}
)

type RunConfig struct {
	Wait bool
}
