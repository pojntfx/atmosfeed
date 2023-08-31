package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"signature"

	"github.com/loopholelabs/scale"
	"github.com/loopholelabs/scale/scalefunc"
)

func main() {
	classifier := flag.String("classifier", filepath.Join("out", "local-everything-latest.scale"), "Path to the classifier Scale function to use")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b, err := os.ReadFile(*classifier)
	if err != nil {
		panic(err)
	}

	f := &scalefunc.Schema{}
	if err := f.Decode(b); err != nil {
		panic(err)
	}

	r, err := scale.New(scale.NewConfig(signature.New).WithFunction(f))
	if err != nil {
		panic(err)
	}

	i, err := r.Instance()
	if err != nil {
		panic(err)
	}

	s := signature.New()
	s.Context.Include = false

	if err := i.Run(ctx, s); err != nil {
		panic(err)
	}

	fmt.Println(s.Context.Include)
}
