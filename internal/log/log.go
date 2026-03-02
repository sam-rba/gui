package log

import (
	"log"
	"os"
)

var (
	Out = log.New(os.Stdout, "faiface/gui [info]: ", log.LstdFlags)
	Err = log.New(os.Stderr, "faiface/gui [error]: ", log.LstdFlags|log.Llongfile)
)
