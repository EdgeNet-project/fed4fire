package utils

import (
	"os"

	"k8s.io/klog/v2"
)

func Check(err error) {
	if err != nil {
		klog.ErrorSDepth(1, err, "Unexpected error")
		os.Exit(1)
	}
}
