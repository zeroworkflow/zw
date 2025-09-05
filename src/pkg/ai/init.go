package ai

import (
	"zero-workflow/src/pkg/ai/zai"
)

func init() {
	// Register Z.ai provider by default
	DefaultFactory.RegisterProvider("z.ai", zai.NewProvider())
}
