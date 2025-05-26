package pal

var (
	defaultConfig = &RunConfig{
		Wait: true,
	}
)

type RunConfig struct {
	Wait bool
}

func runConfigOrDefault(c *RunConfig) *RunConfig {
	if c == nil {
		return defaultConfig
	}
	return c
}
