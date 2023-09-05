module german

go 1.20

require (
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63
	signature v0.1.0
)

require github.com/loopholelabs/polyglot v1.1.2 // indirect

replace signature v0.1.0 => ../../pkg/signatures/classifier/go/guest
